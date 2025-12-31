using System.Text;
using System.Text.Json;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Dispatchers;

public class SlackDispatcher : IChannelDispatcher
{
    private readonly HttpClient _httpClient;
    private readonly string _webhookUrl;

    public SlackDispatcher(HttpClient httpClient, string webhookUrl)
    {
        _httpClient = httpClient;
        _webhookUrl = webhookUrl;
        Name = "Slack";
    }

    public string Name { get; }

    public async Task<(bool Success, string? Error)> SendAsync(NotificationRequest request)
    {
        try
        {
            var payload = new
            {
                text = request.Title,
                blocks = new object[]
                {
                    new
                    {
                        type = "header",
                        text = new
                        {
                            type = "plain_text",
                            text = request.Title
                        }
                    },
                    new
                    {
                        type = "section",
                        text = new
                        {
                            type = "mrkdwn",
                            text = request.Body
                        }
                    },
                    new
                    {
                        type = "context",
                        elements = new object[]
                        {
                            new
                            {
                                type = "mrkdwn",
                                text = $"*Priority:* {request.Priority} | *Service:* {request.Service ?? "N/A"}"
                            }
                        }
                    }
                }
            };

            var json = JsonSerializer.Serialize(payload);
            var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await _httpClient.PostAsync(_webhookUrl, content);
            
            if (response.IsSuccessStatusCode)
            {
                return (true, null);
            }

            var errorContent = await response.Content.ReadAsStringAsync();
            return (false, $"Slack API returned {response.StatusCode}: {errorContent}");
        }
        catch (Exception ex)
        {
            return (false, $"Failed to send Slack notification: {ex.Message}");
        }
    }
}
