# DNS Monitor

**DNS Monitor** is a lightweight Go application that checks for DNS changes related to emails.  
It has been tested and configured specifically for **iCloud emails**, but it can be used for **any email provider**.

## 🚀 Getting Started

### Running the Application
To start the **DNS Monitor**, use:

```bash
docker-compose up --build -d
# To view logs in real-time:
docker logs -f dns-monitor

# To stop and remove the container:
docker-compose down
```

## ⚙️ Configuration Parameters

You can configure the application using **environment variables** before running it.

| Variable            | Description                                  | Required | Default Value             |
|---------------------|----------------------------------------------|----------|---------------------------|
| `DOMAIN`           | The domain to monitor for DNS changes        | ✅ Yes  | _None_                     |
| `PUSHOVER_TOKEN`   | Your Pushover application token              | ✅ Yes  | _None_                     |
| `PUSHOVER_USER`    | Your Pushover user key                       | ✅ Yes  | _None_                     |
| `DNS_SERVER`       | The DNS server to use for queries            | ❌ No   | `1.1.1.1:53` (Cloudflare)  |
| `CHECK_INTERVAL`   | Frequency of DNS checks (`1m`, `10m`, `1h`)  | ❌ No   | `1h`                        |
| `NOTIFY_ON_ERRORS` | Send notifications for application errors    | ❌ No   | `false`                     |
| `CUSTOM_DOMAINS`   | Additional subdomains to monitor (comma-separated) | ❌ No   | _Empty_                     |

### Example Configuration
Before starting the application, export the required variables:

```bash
export DOMAIN="example.com"
export PUSHOVER_TOKEN="your_pushover_app_token"
export PUSHOVER_USER="your_pushover_user_key"
export DNS_SERVER="8.8.8.8:53"
export CHECK_INTERVAL="5m"
export NOTIFY_ON_ERRORS="true"
export CUSTOM_DOMAINS="sub1.example.com,sub2.example.com"
```

Then run:

```bash
docker-compose up --build -d
```

---

### 🎯 Why Use DNS Monitor?
✅ **Lightweight** – Runs efficiently with minimal resources.  
✅ **Flexible** – Can monitor any email domain.  
✅ **Notifications** – Get instant alerts via **Pushover** if changes are detected.  
✅ **Customizable** – Choose the DNS server, check interval, and more!  

---

Enjoy monitoring your DNS! 🚀