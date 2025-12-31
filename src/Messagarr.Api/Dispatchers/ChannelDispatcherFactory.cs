using System.Text.Json;
using Messagarr.Core.Entities;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Dispatchers;

public class ChannelDispatcherFactory
{
    private readonly IHttpClientFactory _httpClientFactory;

    public ChannelDispatcherFactory(IHttpClientFactory httpClientFactory)
    {
        _httpClientFactory = httpClientFactory;
    }

    public IChannelDispatcher CreateDispatcher(Channel channel)
    {
        var config = JsonSerializer.Deserialize<ChannelConfiguration>(channel.Configuration);
        
        if (config == null)
        {
            throw new InvalidOperationException($"Invalid configuration for channel {channel.Name}");
        }

        var httpClient = _httpClientFactory.CreateClient();
        httpClient.Timeout = TimeSpan.FromSeconds(30);

        return channel.Type switch
        {
            ChannelType.Discord => new DiscordDispatcher(httpClient, config.WebhookUrl 
                ?? throw new InvalidOperationException("Discord webhook URL is required")),
            
            ChannelType.Slack => new SlackDispatcher(httpClient, config.WebhookUrl 
                ?? throw new InvalidOperationException("Slack webhook URL is required")),
            
            ChannelType.Teams => new TeamsDispatcher(httpClient, config.WebhookUrl 
                ?? throw new InvalidOperationException("Teams webhook URL is required")),
            
            ChannelType.Email => new EmailDispatcher(
                config.SmtpHost ?? throw new InvalidOperationException("SMTP host is required"),
                config.SmtpPort ?? 587,
                config.SmtpUser ?? throw new InvalidOperationException("SMTP user is required"),
                config.SmtpPassword ?? throw new InvalidOperationException("SMTP password is required"),
                config.FromEmail ?? throw new InvalidOperationException("From email is required"),
                config.ToEmail ?? throw new InvalidOperationException("To email is required")),
            
            _ => throw new NotSupportedException($"Channel type {channel.Type} is not supported")
        };
    }
}
