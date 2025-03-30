# DNS Monitor – Real-Time DNS Change Detection & Email Security

**DNS Monitor** is a lightweight and customizable **DNS change detection tool** designed to enhance **email security** and prevent unauthorized modifications to critical **DNS records** like **MX, SPF, DKIM, and DMARC**. It provides instant notifications via **Pushover** and **Telegram**, ensuring you stay informed of potential threats. While it was specifically tested and configured for **iCloud emails**, it can be used for **any email provider** or **general DNS monitoring**.

✅ **Monitor DNS changes in real-time**  
✅ **Protect against email hijacking, spoofing, and phishing**  
✅ **Get instant alerts when records are modified**  
✅ **Flexible & customizable for any domain**

## Why Monitor DNS Records?  
Email security relies on key DNS records to prevent hijacking, spoofing, and phishing. Unauthorized modifications can compromise your email integrity. **DNS Monitor** helps you detect and respond to such changes in real time by tracking:  

- **MX (Mail Exchange):** Controls email routing; changes can hijack mail flow.  
- **SPF (Sender Policy Framework):** Defines authorized email senders; unauthorized changes allow spoofing.  
- **DKIM (DomainKeys Identified Mail):** Cryptographically signs emails; modifications enable forged emails.  
- **DMARC (Domain-based Message Authentication, Reporting, and Conformance):** Enforces email authentication policies; weakening it allows phishing attempts.  

Enable alerts to detect unauthorized modifications and maintain secure email communication.  


## 🚀 Getting Started

### Running the Application 

#### **1️⃣ Run with Environment Variables**
```bash
# Pushover example
docker run -d --name dns-monitor \
  -e DOMAIN="example.com" \
  -e NOTIFIER_TYPE="pushover" \
  -e PUSHOVER_APP_TOKEN="your-token" \
  -e PUSHOVER_USER_KEY="your-user-key" \
  1gabriel/dns-monitor:latest
# Telegram example
docker run -d --name dns-monitor \
  -e DOMAIN="example.com" \
  -e NOTIFIER_TYPE="telegram" \
  -e TELEGRAM_BOT_TOKEN="your-bot-token" \
  -e TELEGRAM_CHAT_IDS="your-chat-id" \
  1gabriel/dns-monitor:latest

docker logs dns-monitor
```

#### **2️⃣ Run with Docker Compose**
```bash
# Run by Pulling the Remote Image
docker-compose up -d
docker logs dns-monitor
```

### Development Mode
To start the **DNS Monitor**, use:

```bash
git clone git@github.com:xegabriel/dns-monitor.git
cd dns-monitor
# Run by Building Local Files
docker-compose up --build -d
# To view logs in real-time:
docker logs -f dns-monitor

# To stop and remove the container:
docker-compose down

# Without docker
go run cmd/main.go

# Run the tests
go test ./...
```

## ⚙️ Configuration Parameters

| Variable                | Description                                           | Required | Default Value              |
|-------------------------|-------------------------------------------------------|----------|----------------------------|
| `DOMAIN`               | The domain to monitor for DNS changes                 | ✅ Yes  | _None_                      |
| `NOTIFIER_TYPE`        | Notification method (`pushover` or `telegram`)        | ✅ Yes  | _None_                      |
| `PUSHOVER_APP_TOKEN`   | Pushover application token (Required if using Pushover)  | ✅* Yes | _None_                      |
| `PUSHOVER_USER_KEY`    | Pushover user key (Required if using Pushover)        | ✅* Yes | _None_                      |
| `TELEGRAM_BOT_TOKEN`   | Telegram bot token (Required if using Telegram)       | ✅* Yes | _None_                      |
| `TELEGRAM_CHAT_IDS`    | Comma-separated list of Telegram chat IDs (Required if using Telegram) | ✅* Yes | _None_  |
| `DNS_SERVER`           | The DNS server to use for queries                     | ❌ No   | `1.1.1.1:53` (Cloudflare)   |
| `CHECK_INTERVAL`       | Frequency of DNS checks (`1m`, `10m`, `1h`)           | ❌ No   | `1h`                         |
| `NOTIFY_ON_ERRORS`     | Send notifications for application errors             | ❌ No   | `false`                      |
| `CUSTOM_SUBDOMAINS`    | Additional subdomains to monitor (comma-separated)    | ❌ No   | _Empty_                      |
| `CUSTOM_DKIM_SELECTORS`| Additional DKIM selectors to monitor (comma-separated). Check the `DKIM Selectors` section for examples | ❌ No   | _Empty_                      |

> **Note:**  
> - `PUSHOVER_APP_TOKEN` and `PUSHOVER_USER_KEY` are required **only if** `NOTIFIER_TYPE=pushover`.  
> - `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_IDS` are required **only if** `NOTIFIER_TYPE=telegram`.  
> - Only one notifier type can be used at a time.  

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
export NOTIFIER_TYPE="pushover"
export PUSHOVER_APP_TOKEN="your_pushover_app_token"
export PUSHOVER_USER_KEY="your_pushover_user_key"
export DNS_SERVER="8.8.8.8:53"
export CHECK_INTERVAL="5m"
export NOTIFY_ON_ERRORS="true"
export CUSTOM_SUBDOMAINS="sub1,sub2"
export CUSTOM_DKIM_SELECTORS="*,sig1"
```

## 🔔 Notification Integration  

**DNS Monitor** uses [`nikoksr/notify`](https://github.com/nikoksr/notify) to integrate multiple notification services. This makes it easy to extend support for additional services as needed.  

### Supported Notification Services  

At this moment, **DNS Monitor** supports the following notification services:  
- **Pushover**  
- **Telegram**  

### Adding a New Notification Service  

To add a new notification service, follow these steps:  

1. **Add a new provider**: Implement the provider in [`internal/notification/providers`](https://github.com/xegabriel/dns-monitor/tree/main/internal/notification/providers).  
2. **Register the provider in the factory**: Modify [`factory.go`](https://github.com/xegabriel/dns-monitor/blob/main/internal/notification/factory.go#L23) to include the new provider.  
3. **Update allowed notifier types**: Add the new notifier type in [`types.go`](https://github.com/xegabriel/dns-monitor/blob/main/internal/common/types.go#L45).  
4. **Load the required environment variables**: Modify [`config.go`](https://github.com/xegabriel/dns-monitor/blob/main/internal/common/config.go#L104) to support the new service's configuration.  
5. **Update the Docker Compose file**: Add necessary environment variables in [`docker-compose.yml`](https://github.com/xegabriel/dns-monitor/blob/main/docker-compose.yml).  
6. **Update this README**: Document the new notifier under the **Configuration Parameters** section.  

By following these steps, you can seamlessly integrate new notification services into **DNS Monitor**. 🚀  

---

## 📜 Disclaimer
The authors of this library bear no responsibility for any misuse or unintended consequences arising from its use. Users assume full liability for their actions. For more details, refer to the [LICENSE](https://github.com/xegabriel/dns-monitor/blob/main/LICENSE).

---

## ⭐ Support
If you find this project helpful, please consider giving it a ⭐️ on GitHub!

Enjoy monitoring your DNS! 🚀