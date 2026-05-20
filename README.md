# HA Add-on for Automatic mDNS

## Build

```bash
# Using Docker
docker build -t ha-addon-auto-mdns:latest -f addon/Dockerfile .

# Using Bazel during testing
bazel build //cmd/auto-mdns:auto-mdns
```

## Publish to GitHub Container Registry (GHCR)

This add-on is published to [ghcr.io/chickenandpork/ha-addon-auto-mdns](https://ghcr.io/chickenandpork/ha-addon-auto-mdns) via GitHub Actions on every merge to the default branch.

### Manual publish

```bash
# Build and tag for GHCR
IMAGE=ghcr.io/chickenandpork/ha-addon-auto-mdns
VERSION=$(cat addon/VERSION)
docker build -t ${IMAGE}:${VERSION} -t ${IMAGE}:latest -f addon/Dockerfile .

# Log in to GHCR
docker login ghcr.io -u <your-username> -p <personal-access-token>

# Push the image
docker push ${IMAGE}:${VERSION}
docker push ${IMAGE}:latest
```

> **Note:** Use a [personal access token (PAT)](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) with `write:packages` scope when authenticating to GHCR.

## Install the Add-on in Home Assistant

### 1. Add the repository to Home Assistant

The Home Assistant add-on store reads from a repository index file. To add this add-on:

1. In Home Assistant, go to **Supervisor** > **Add-on Store** > **Repositories** (three-dot menu).
2. Click **Add Repository** and enter the URL to this repository's `repositories.json`:
   ```
   https://raw.githubusercontent.com/chickenandpork/ha-addon-auto-mdns/main/repositories.json
   ```
3. Click **Submit**. The add-on "HA Add-on Auto-mDNS" will appear in the store.

### 2. Install and configure the add-on

1. Click **HA Add-on Auto-mDNS** in the store to open its details page.
2. Click **Install**.
3. Configure the options in the **Configuration** tab:

   | Option              | Default        | Description                          |
   |---------------------|----------------|--------------------------------------|
   | `listen_address`    | `:80`          | HTTP proxy listen address            |
   | `https_listen_address` | `:443`     | HTTPS proxy listen address           |
   | `poll_interval`     | `30s`          | How often to poll for addon changes  |
   | `local_hostname`    | `auto-mdns`    | Local mDNS hostname (e.g. `auto-mdns.local`) |

4. Click **Save** and then **Start**.

### 3. Verify mDNS discovery

After starting the add-on, addon services should be discoverable via mDNS by their slug names:

```bash
# Example: discover the proxy service via mDNS
dns-sd -B _http._tcp

# Example: resolve a specific addon's mDNS name
dns-sd -G v4 my-addon.local
```

## Development

### Build the binary

```bash
# Using Bazel (recommended)
bazel build //cmd/auto-mdns:auto-mdns
bazel-bin/cmd/auto-mdns/auto-mdns
```

### Run tests

```bash
bazel test //...
```

### Run gazelle

```bash
bazel run //:gazelle -- update
```

## GitHub Actions CI/CD

Merges to the default branch trigger a GitHub Actions workflow that:

1. Builds the Docker image with Bazel
2. Publishes it to `ghcr.io/chickenandpork/ha-addon-auto-mdns`

The workflow is defined in [`.github/workflows/publish.yaml`](.github/workflows/publish.yaml).

### Required secrets

| Secret              | Description                                    |
|---------------------|------------------------------------------------|
| `GHCR_TOKEN`        | GitHub PAT with `write:packages` scope         |

Set this in your repository's **Settings > Secrets and variables > Actions**.
