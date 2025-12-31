using System.Text;
using System.Text.Json;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Dispatchers;

public class TeamsDispatcher : IChannelDispatcher
{
    private readonly HttpClient _httpClient;
    private readonly string _webhookUrl;

    public TeamsDispatcher(HttpClient httpClient, string webhookUrl)
    {
        _httpClient = httpClient;
        _webhookUrl = webhookUrl;
        Name = "Teams";
    }

    public string Name { get; }

    public async Task<(bool Success, string? Error)> SendAsync(NotificationRequest request)
    {
        try
        {
            var payload = new
            {
                type = "message",
                attachments = new object[]
                {
                    new
                    {
                        contentType = "application/vnd.microsoft.card.adaptive",
                        content = new
                        {
                            type = "AdaptiveCard",
                            body = new object[]
                            {
                                new
                                {
                                    type = "TextBlock",
                                    size = "Large",
                                    weight = "Bolder",
                                    text = request.Title
                                },
                                new
                                {
                                    type = "TextBlock",
                                    text = request.Body,
                                    wrap = true
                                },
                                new
                                {
                                    type = "FactSet",
                                    facts = new object[]
                                    {
                                        new { title = "Priority", value = request.Priority },
                                        new { title = "Service", value = request.Service ?? "N/A" }
                                    }
                                }
                            },
                            schema = "http://adaptivecards.io/schemas/adaptive-card.json",
                            version = "1.4"
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
            return (false, $"Teams API returned {response.StatusCode}: {errorContent}");
        }
        catch (Exception ex)
        {
            return (false, $"Failed to send Teams notification: {ex.Message}");
        }
    }
}
