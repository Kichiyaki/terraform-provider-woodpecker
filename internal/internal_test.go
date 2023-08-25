package internal_test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	urlpkg "net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
	"golang.org/x/oauth2"
)

const resourceExpInSec = 120

var (
	giteaClient      *gitea.Client
	woodpeckerClient woodpecker.Client
)

func TestMain(m *testing.M) {
	pool := newDockertestPool()

	var deferFuncs []func()

	network := newDockerNetwork(pool)
	deferFuncs = append(deferFuncs, func() {
		_ = network.Close()
	})

	resourceGitea := runGitea(pool, network)
	deferFuncs = append(deferFuncs, func() {
		_ = resourceGitea.Close()
	})

	giteaClient = newGiteaClient(resourceGitea.httpURL, resourceGitea.user)

	resourceWoodpecker := runWoodpecker(pool, network, resourceGitea.httpURL, resourceGitea.privateHTTPURL, resourceGitea.user)
	deferFuncs = append(deferFuncs, func() {
		_ = resourceWoodpecker.Close()
	})

	// set required envs
	_ = os.Setenv("TF_ACC", "1")
	_ = os.Setenv("WOODPECKER_SERVER", resourceWoodpecker.httpURL.String())
	_ = os.Setenv("WOODPECKER_TOKEN", resourceWoodpecker.token)

	woodpeckerClient = newWoodpeckerClient(resourceWoodpecker.httpURL, resourceWoodpecker.token)

	code := m.Run()

	// LIFO
	for i := len(deferFuncs) - 1; i >= 0; i-- {
		deferFuncs[i]()
	}

	os.Exit(code)

}

func newDockertestPool() *dockertest.Pool {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("couldn't construct pool: %s", err)
	}

	if err = pool.Client.Ping(); err != nil {
		log.Fatalf("couldn't connect to Docker: %s", err)
	}

	return pool
}

func newDockerNetwork(pool *dockertest.Pool) *dockertest.Network {
	network, err := pool.CreateNetwork(uuid.NewString())
	if err != nil {
		log.Fatalln("couldn't create docker network:", err)
	}
	return network
}

type giteaResource struct {
	docker         *dockertest.Resource
	httpURL        *urlpkg.URL
	privateHTTPURL *urlpkg.URL
	user           *urlpkg.Userinfo
}

func runGitea(pool *dockertest.Pool, network *dockertest.Network) giteaResource {
	gitea, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "gitea/gitea",
		Tag:        "1.20",
		Networks:   []*dockertest.Network{network},
		Env: []string{
			"GITEA__security__INSTALL_LOCK=true",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("couldn't run gitea: %s", err)
	}

	if err = gitea.Expire(resourceExpInSec); err != nil {
		log.Fatal(err)
	}

	httpURL := &urlpkg.URL{
		Scheme: "http",
		Host:   getHostPort(gitea, "3000/tcp"),
	}

	if err = pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, httpURL.String(), nil)
		if reqErr != nil {
			return reqErr
		}

		res, resErr := (&http.Client{}).Do(req)
		if resErr != nil {
			return resErr
		}
		defer func() {
			_ = res.Body.Close()
		}()

		_, _ = io.Copy(io.Discard, res.Body)

		if res.StatusCode != http.StatusOK {
			return errors.New("request to gitea failed")
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	return giteaResource{
		docker:  gitea,
		httpURL: httpURL,
		privateHTTPURL: &urlpkg.URL{
			Scheme: httpURL.Scheme,
			Host:   gitea.GetIPInNetwork(network) + ":3000",
		},
		user: createGiteaUser(pool, gitea),
	}
}

func (r giteaResource) Close() error {
	return r.docker.Close()
}

func createGiteaUser(pool *dockertest.Pool, gitea *dockertest.Resource) *urlpkg.Userinfo {
	username := strings.ReplaceAll(uuid.NewString(), "-", "")
	password := uuid.NewString()

	stdOutBuf := bytes.NewBuffer(nil)
	stdErrBuf := bytes.NewBuffer(nil)

	exec, err := pool.Client.CreateExec(docker.CreateExecOptions{
		Container: gitea.Container.ID,
		User:      "git",
		Cmd: []string{
			"gitea",
			"admin",
			"user",
			"create",
			"--admin",
			"--username=" + username,
			"--password=" + password,
			fmt.Sprintf("--email=%s@localhost", username),
		},
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		log.Fatalln("couldn't create exec:", err)
	}

	err = pool.Client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: stdOutBuf,
		ErrorStream:  stdErrBuf,
	})
	if err != nil {
		log.Fatalln("couldn't start exec:", err)
	}

	inspectExec, err := pool.Client.InspectExec(exec.ID)
	if err != nil {
		log.Fatalln("couldn't inspect exec:", err)
	}

	if inspectExec.ExitCode != 0 {
		log.Fatalln("couldn't create user\nstdout:", stdOutBuf.String(), "\nstderr:", stdErrBuf.String())
	}

	return urlpkg.UserPassword(username, password)
}

type woodpeckerResource struct {
	docker  *dockertest.Resource
	httpURL *urlpkg.URL
	token   string
}

func runWoodpecker(
	pool *dockertest.Pool,
	network *dockertest.Network,
	giteaPublicURL, giteaPrivateURL *urlpkg.URL,
	giteaUser *urlpkg.Userinfo,
) woodpeckerResource {
	httpURL := &urlpkg.URL{
		Scheme: giteaPublicURL.Scheme,
		//nolint:gosec
		Host: giteaPublicURL.Hostname() + ":" + strconv.Itoa(rand.New(rand.NewSource(time.Now().Unix())).Intn(5000)+35000),
	}

	oauthApp, _, err := giteaClient.CreateOauth2(gitea.CreateOauth2Option{
		Name:               "woodpecker",
		ConfidentialClient: true,
		RedirectURIs:       []string{httpURL.String() + "/authorize"},
	})
	if err != nil {
		log.Fatalln("couldn't create oauth2 app:", err)
	}

	woodpecker, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "woodpeckerci/woodpecker-server",
		Tag:        "v1.0.2",
		Networks:   []*dockertest.Network{network},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"8000/tcp": {
				{
					HostPort: httpURL.Port(),
				},
			},
		},
		Env: []string{
			"WOODPECKER_OPEN=true",
			"WOODPECKER_HOST=" + httpURL.String(),
			"WOODPECKER_AGENT_SECRET=" + uuid.NewString(),
			"WOODPECKER_GITEA=true",
			"WOODPECKER_GITEA_URL=" + giteaPrivateURL.String(),
			"WOODPECKER_GITEA_CLIENT=" + oauthApp.ClientID,
			"WOODPECKER_GITEA_SECRET=" + oauthApp.ClientSecret,
			"WOODPECKER_ADMIN=" + giteaUser.Username(),
			"WOODPECKER_LOG_LEVEL=debug",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.Mounts = append(config.Mounts, docker.HostMount{
			Type:   "tmpfs",
			Target: "/var/lib/woodpecker",
		})
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("couldn't run woodpecker: %s", err)
	}

	if err = woodpecker.Expire(resourceExpInSec); err != nil {
		log.Fatal(err)
	}

	if err = pool.Retry(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, httpURL.String(), nil)
		if reqErr != nil {
			return reqErr
		}

		res, resErr := (&http.Client{}).Do(req)
		if resErr != nil {
			return resErr
		}
		defer func() {
			_ = res.Body.Close()
		}()

		_, _ = io.Copy(io.Discard, res.Body)

		if res.StatusCode != http.StatusOK {
			return errors.New("request to woodpecker failed")
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	return woodpeckerResource{
		docker:  woodpecker,
		httpURL: httpURL,
		token:   newWoodpeckerTokenProvider(oauthApp, giteaUser, giteaPublicURL, httpURL).token(),
	}
}

func (r woodpeckerResource) Close() error {
	return r.docker.Close()
}

type woodpeckerTokenProvider struct {
	client        *http.Client
	oauthApp      *gitea.Oauth2
	giteaUser     *urlpkg.Userinfo
	giteaURL      *urlpkg.URL
	woodpeckerURL *urlpkg.URL
}

func newWoodpeckerTokenProvider(
	oauthApp *gitea.Oauth2,
	giteaUser *urlpkg.Userinfo,
	giteaURL *urlpkg.URL,
	woodpeckerURL *urlpkg.URL,
) woodpeckerTokenProvider {
	cookieJar, _ := cookiejar.New(nil)
	return woodpeckerTokenProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Jar:     cookieJar,
		},
		oauthApp:      oauthApp,
		giteaUser:     giteaUser,
		giteaURL:      giteaURL,
		woodpeckerURL: woodpeckerURL,
	}
}

func (p woodpeckerTokenProvider) token() string {
	ctx := context.Background()

	// we need to make this request to get the csrf token
	respGetLoginPage := p.get(ctx, p.giteaURL.String()+"/user/login")
	_, _ = io.Copy(io.Discard, respGetLoginPage.Body)
	_ = respGetLoginPage.Body.Close()

	// log in to Gitea
	password, _ := p.giteaUser.Password()
	respLogin := p.postForm(ctx, p.giteaURL.String()+"/user/login", urlpkg.Values{
		"_csrf":     []string{p.getCSRFTokenFromCookies(p.client.Jar.Cookies(p.giteaURL))},
		"user_name": []string{p.giteaUser.Username()},
		"password":  []string{password},
	})
	_, _ = io.Copy(io.Discard, respLogin.Body)
	_ = respLogin.Body.Close()

	// we need to make this request to get the csrf token
	respAuthorize := p.get(ctx, (&urlpkg.URL{
		Scheme: p.giteaURL.Scheme,
		Host:   p.giteaURL.Host,
		Path:   "/login/oauth/authorize",
		RawQuery: urlpkg.Values{
			"client_id":     []string{p.oauthApp.ClientID},
			"redirect_uri":  []string{p.oauthApp.RedirectURIs[0]},
			"response_type": []string{"code"},
			"state":         []string{p.oauthApp.Name},
		}.Encode(),
	}).String())
	_, _ = io.Copy(io.Discard, respAuthorize.Body)
	_ = respAuthorize.Body.Close()

	// log in to Woodpecker
	respGrant := p.postForm(ctx, p.giteaURL.String()+"/login/oauth/grant", urlpkg.Values{
		"_csrf":         []string{p.getCSRFTokenFromCookies(p.client.Jar.Cookies(p.giteaURL))},
		"client_id":     []string{p.oauthApp.ClientID},
		"redirect_uri":  []string{p.oauthApp.RedirectURIs[0]},
		"state":         []string{p.oauthApp.Name},
		"response_type": []string{"code"},
		"scope":         []string{""},
		"nonce":         []string{""},
	})
	_, _ = io.Copy(io.Discard, respGrant.Body)
	_ = respGrant.Body.Close()

	// we need to make this request to get the csrf token
	respWebConfig := p.get(ctx, p.woodpeckerURL.String()+"/web-config.js")
	csrfToken := p.readCSRFTokenFromWoodpeckerWebConfig(respWebConfig.Body)
	_ = respWebConfig.Body.Close()
	if csrfToken == "" {
		log.Fatalln("couldn't extract csrf token from woodpecker web config")
	}

	// finally generate the woodpecker token
	reqToken := p.newRequestWithContext(ctx, http.MethodPost, p.woodpeckerURL.String()+"/api/user/token", nil)
	reqToken.Header.Set("X-Csrf-Token", csrfToken)

	respToken := p.do(reqToken)
	token, err := io.ReadAll(respToken.Body)
	if err != nil {
		log.Fatalln("couldn't read token from response:", err)
	}
	_ = respToken.Body.Close()

	return string(token)
}

func (p woodpeckerTokenProvider) postForm(ctx context.Context, url string, values urlpkg.Values) *http.Response {
	req := p.newRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return p.do(req)
}

func (p woodpeckerTokenProvider) get(ctx context.Context, url string) *http.Response {
	return p.do(p.newRequestWithContext(ctx, http.MethodGet, url, nil))
}

func (p woodpeckerTokenProvider) newRequestWithContext(ctx context.Context, method string, url string, body io.Reader) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		log.Fatalf("couldn't construct request for url %s: %s", url, err)
	}

	return req
}

func (p woodpeckerTokenProvider) do(req *http.Request) *http.Response {
	resp, err := p.client.Do(req)
	if err != nil {
		log.Fatalf("request to %s failed: %s", req.URL.String(), err)
	}
	if resp.StatusCode/100 != 2 { // accept only 2XX requests
		log.Fatalf("request to %s failed", req.URL.String())
	}

	return resp
}

func (p woodpeckerTokenProvider) getCSRFTokenFromCookies(cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "_csrf" {
			return cookie.Value
		}
	}
	return ""
}

func (p woodpeckerTokenProvider) readCSRFTokenFromWoodpeckerWebConfig(r io.Reader) string {
	defer func() {
		// discard the remaining bytes
		_, _ = io.Copy(io.Discard, r)
	}()

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !strings.HasPrefix(line, "window.WOODPECKER_CSRF") {
			continue
		}

		return line[strings.Index(line, `"`)+1 : strings.LastIndex(line, `"`)]
	}

	return ""
}

func getHostPort(resource *dockertest.Resource, id string) string {
	dockerURL := os.Getenv("DOCKER_HOST")
	if dockerURL == "" {
		return resource.GetHostPort(id)
	}

	u, err := urlpkg.Parse(dockerURL)
	if err != nil {
		log.Fatalln("couldn't parse DOCKER_HOST:", err)
	}

	return u.Hostname() + ":" + resource.GetPort(id)
}

func newGiteaClient(url *urlpkg.URL, user *urlpkg.Userinfo) *gitea.Client {
	password, _ := user.Password()
	client, err := gitea.NewClient(
		url.String(),
		gitea.SetHTTPClient(&http.Client{
			Timeout: 10 * time.Second,
		}),
		gitea.SetBasicAuth(user.Username(), password),
	)
	if err != nil {
		log.Fatalln("couldn't construct *gitea.Client:", err)
	}

	if _, _, err = client.GetMyUserInfo(); err != nil {
		log.Fatalln("couldn't get user info:", err)
	}

	return client
}

func newWoodpeckerClient(url *urlpkg.URL, token string) woodpecker.Client {
	client := woodpecker.NewClient(
		url.String(),
		(&oauth2.Config{}).Client(context.Background(), &oauth2.Token{
			AccessToken: token,
		}),
	)

	if _, err := client.Self(); err != nil {
		log.Fatalln("couldn't get user info:", err)
	}

	return client
}
