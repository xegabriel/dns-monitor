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
| `CUSTOM_SUBDOMAINS`   | Additional subdomains to monitor (comma-separated) | ❌ No   | _Empty_                     |
| `CUSTOM_DKIM_SELECTORS`   | Additional DKIM selectors to monitor (comma-separated) | ❌ No   | _Empty_                     |

### DKIM Selectors

| **Email Provider** | **Common DKIM Selectors** | **Example DKIM Record** |
|--------------------|-------------------------|--------------------------|
| **Google (Gmail, Google Workspace)** | `google`, `default` | `google._domainkey.example.com` |
| **iCloud (Apple Mail)** | `sig1` | `sig1._domainkey.example.com` |
| **Microsoft (Outlook, Office 365, Exchange)** | `selector1`, `selector2` | `selector1._domainkey.example.com` |
| **Yahoo! Mail** | `selector1`, `selector2` | `selector1._domainkey.example.com` |
| **Zoho Mail** | `zoho` | `zoho._domainkey.example.com` |
| **Proton Mail** | `protonmail1`, `protonmail2` | `protonmail1._domainkey.example.com` |
| **FastMail** | `fm1`, `fm2` | `fm1._domainkey.example.com` |
| **Amazon SES** | `amazon`, `selector1`, `selector2` | `selector1._domainkey.example.com` |

### Notes:
- Some providers may generate **custom DKIM selectors** for each domain.
- To check your **exact DKIM selector**, inspect your existing DNS records to verify the active selectors in use.
  
### Monitoring DKIM:
It’s a good idea to monitor your DKIM records for any unexpected changes, as altering these can affect email authenticity, security, and deliverability. Unauthorized changes to DKIM selectors could indicate a **compromise** or a misconfiguration in your email system. Regular audits can help identify potential vulnerabilities.


### Example Configuration
Before starting the application, export the required variables:

```bash
export DOMAIN="example.com"
export PUSHOVER_TOKEN="your_pushover_app_token"
export PUSHOVER_USER="your_pushover_user_key"
export DNS_SERVER="8.8.8.8:53"
export CHECK_INTERVAL="5m"
export NOTIFY_ON_ERRORS="true"
export CUSTOM_SUBDOMAINS="sub1.example.com,sub2.example.com"
export CUSTOM_DKIM_SELECTORS="*,sig1"
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