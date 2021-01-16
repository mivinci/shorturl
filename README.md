![shorturl](https://socialify.git.ci/mivinci/shorturl/image?description=1&language=1&owner=1&stargazers=1&theme=Light)

## Try it out

Demo site: [https://u.xjj.pub](https://l.xjj.pub)

Or you may wanna offer me a nicer and shorter domain 🙃.

## Deploy on your own host

Apparently deploying by [docker](https://www.docker.com/) is suggested, so make sure you have docker n your host before everything.

- **Clone this repo**

```bash
git clone git@github.com:Mivinci/shorturl.git
cd shorturl
```

- **Build the docker image**

```bash
docker build -t shorturl .
```

- **Run the image**

```bash
docker run --rm -d -p 5000:8080 shorturl
```

Surely, you can change `5000` to any other port you'd like to deploy on for this  project.

- **Use Caddy for reverse proxy (Optional)**

Make sure you have installed [Caddy](https://caddyserver.com) on your host.

1. Create a `Caddyfile` in the root directory of the repo and write the following content in it. 

```nginx
xxx.com {
    reverse_proxy localhost:5000
}
```

You need to change `xxx.com` to your own domain name.

2. Run the following commands to activate the Caddyfile you just wrote and reload Caddy.

```bash
caddy adapt
caddy reload
```

Besides, Caddy will automatically generate certificate files for your site, so you can visit it by `https` after using Caddy.

## About this repo

I didn't use `go mod`, also known as `go module` to manage this project, because I wanted to keep the root directory clean. It doesn't actually bother me not to use `go mod` for such a small project. Also, I'd like to regard this project as a minimal implementation for a URL shortening service, so only the most needed functionalities were coded.