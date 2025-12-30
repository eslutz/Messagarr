# Channel Setup Guides

This guide will help you configure each notification channel in Messagarr.

## Table of Contents

- [Email (SMTP)](#email-smtp)
- [Discord](#discord)
- [Slack](#slack)
- [Microsoft Teams](#microsoft-teams)

## Email (SMTP)

### Gmail Setup

1. **Enable 2-Factor Authentication** (if not already enabled)
   - Go to [Google Account Security](https://myaccount.google.com/security)
   - Enable 2-Step Verification

2. **Generate App Password**
   - Go to [App Passwords](https://myaccount.google.com/apppasswords)
   - Select "Mail" and your device
   - Copy the generated 16-character password

3. **Configuration**

   ```yaml
   channels:
     - type: smtp
       name: email
       host: smtp.gmail.com
       port: 587
       user: your-email@gmail.com
       password: ${SMTP_PASSWORD}  # Use app password
       from: your-email@gmail.com
       to: recipient@example.com
   ```

4. **Environment Variables**

   ```bash
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=your-email@gmail.com
   SMTP_PASSWORD=your-app-password
   SMTP_FROM=your-email@gmail.com
   EMAIL_TO=recipient@example.com
   ```

### Other SMTP Providers

#### Outlook/Office 365

```yaml
host: smtp.office365.com
port: 587
```

#### Yahoo Mail

```yaml
host: smtp.mail.yahoo.com
port: 587
```

#### Custom SMTP Server

```yaml
host: mail.example.com
port: 587  # or 465 for SSL, 25 for non-encrypted
```

### Testing

```bash
curl -X POST http://localhost:4545/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Email Test",
    "body": "Testing email notifications from Messagarr",
    "channels": ["email"]
  }'
```

---

## Discord

### Setup

1. **Navigate to Server Settings**
   - Right-click on your Discord server
   - Select "Server Settings"

2. **Create Webhook**
   - Go to "Integrations" → "Webhooks"
   - Click "New Webhook"
   - Give it a name (e.g., "Messagarr")
   - Select the channel where notifications should appear
   - Copy the webhook URL

3. **Configuration**

   ```yaml
   channels:
     - type: discord
       webhook_url: ${DISCORD_WEBHOOK_URL}
   ```

4. **Environment Variables**

   ```bash
   DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/123456789/abcdefghijklmnopqrstuvwxyz
   ```

### Features

- Rich embeds with colors based on priority
- Custom fields for metadata
- Automatic formatting
- Thumbnail and image support (future)

### Testing

```bash
curl -X POST http://localhost:4545/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Discord Test",
    "body": "Testing Discord notifications from Messagarr",
    "priority": "high",
    "metadata": {
      "server": "production",
      "app_version": "1.2.3"
    },
    "channels": ["discord"]
  }'
```

### Example Output

![Discord Example](https://via.placeholder.com/500x300?text=Discord+Notification+Example)

---

## Slack

### Setup

1. **Create Incoming Webhook**
   - Go to [Slack API Apps](https://api.slack.com/apps)
   - Click "Create New App" → "From scratch"
   - Name your app (e.g., "Messagarr")
   - Select your workspace

2. **Enable Incoming Webhooks**
   - In the left sidebar, click "Incoming Webhooks"
   - Toggle "Activate Incoming Webhooks" to On
   - Click "Add New Webhook to Workspace"
   - Select the channel where notifications should appear
   - Click "Allow"

3. **Copy Webhook URL**
   - Copy the webhook URL (starts with `https://hooks.slack.com/services/`)

4. **Configuration**

   ```yaml
   channels:
     - type: slack
       webhook_url: ${SLACK_WEBHOOK_URL}
   ```

5. **Environment Variables**

   ```bash
   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX
   ```

### Features

- Block Kit formatted messages
- Field sections for metadata
- Markdown support
- Emoji and mentions (future)

### Testing

```bash
curl -X POST http://localhost:4545/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Slack Test",
    "body": "Testing Slack notifications from Messagarr",
    "service": "test-service",
    "event_type": "test.notification",
    "channels": ["slack"]
  }'
```

### Example Output

![Slack Example](https://via.placeholder.com/500x300?text=Slack+Notification+Example)

---

## Microsoft Teams

### Setup

1. **Navigate to Channel**
   - Open Microsoft Teams
   - Navigate to the team and channel where you want notifications

2. **Add Incoming Webhook Connector**
   - Click the "..." (more options) next to the channel name
   - Select "Connectors"
   - Search for "Incoming Webhook"
   - Click "Configure"

3. **Configure Webhook**
   - Give it a name (e.g., "Messagarr")
   - Optionally upload an image
   - Click "Create"
   - Copy the webhook URL

4. **Configuration**

   ```yaml
   channels:
     - type: teams
       webhook_url: ${TEAMS_WEBHOOK_URL}
   ```

5. **Environment Variables**

   ```bash
   TEAMS_WEBHOOK_URL=https://outlook.office.com/webhook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx@xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/IncomingWebhook/yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy/zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz
   ```

### Features

- Adaptive Cards format
- Fact sets for structured data
- Rich formatting
- Action buttons (future)

### Testing

```bash
curl -X POST http://localhost:4545/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Teams Test",
    "body": "Testing Microsoft Teams notifications from Messagarr",
    "priority": "normal",
    "service": "messagarr",
    "event_type": "test.notification",
    "metadata": {
      "environment": "development",
      "timestamp": "2025-12-30T01:00:00Z"
    },
    "channels": ["teams"]
  }'
```

### Example Output

![Teams Example](https://via.placeholder.com/500x300?text=Teams+Notification+Example)

---

## Priority Groups

Instead of specifying channels explicitly, you can use priority groups to automatically route notifications:

```yaml
priority_groups:
  high:
    - email
    - discord
  normal:
    - discord
    - slack
  low:
    - slack
```

Then send notifications with priority:

```bash
curl -X POST http://localhost:4545/notify \
  -H "Content-Type: application/json" \
  -d '{
    "title": "High Priority Alert",
    "body": "Critical system issue detected",
    "priority": "high"
  }'
```

This will send to both `email` and `discord` channels.

## Troubleshooting

### All Channels

**Problem**: Notifications not being sent

**Solutions**:

- Check logs: `docker-compose logs messagarr`
- Verify environment variables are set
- Test webhook URLs manually
- Check firewall/network settings

### Email

**Problem**: Authentication failed

**Solutions**:

- Verify username/password
- Use app password instead of account password
- Check 2FA settings
- Try port 465 with SSL

**Problem**: Connection timeout

**Solutions**:

- Check SMTP host and port
- Verify firewall allows outbound SMTP
- Try different ports (25, 465, 587)

### Discord

**Problem**: Webhook not found

**Solutions**:

- Regenerate webhook in Discord
- Ensure webhook URL is complete
- Check channel permissions

**Problem**: Rate limited

**Solutions**:

- Reduce notification frequency
- Increase `dedup_ttl` in config
- Check Discord rate limits (5 requests/2 seconds)

### Slack

**Problem**: "no_service" error

**Solutions**:

- Recreate webhook
- Check workspace permissions
- Verify app is installed in workspace

**Problem**: "channel_not_found"

**Solutions**:

- Reinstall webhook in correct channel
- Check channel still exists
- Verify app has channel access

### Teams

**Problem**: Webhook expired

**Solutions**:

- Webhooks expire after 90 days of inactivity
- Regenerate webhook in Teams
- Update environment variable

**Problem**: "invalid payload"

**Solutions**:

- Check Messagarr logs for validation errors
- Verify adaptive card format
- Test with simple notification first

## Best Practices

1. **Use Priority Groups** - Define clear routing rules
2. **Test Webhooks** - Verify each channel after configuration
3. **Monitor Logs** - Watch for delivery failures
4. **Set Dedup TTL** - Prevent duplicate notifications
5. **Rotate Secrets** - Periodically regenerate webhooks and passwords
6. **Document Setup** - Keep notes on channel configurations
7. **Use Separate Channels** - Different channels for different purposes

## Getting Help

- GitHub Issues: [eslutz/Messagarr/issues](https://github.com/eslutz/Messagarr/issues)
- Documentation: [docs/](../docs/)
- Examples: Check `examples/` directory
