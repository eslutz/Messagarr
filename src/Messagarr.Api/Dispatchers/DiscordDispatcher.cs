using System.Text;
using System.Text.Json;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Dispatchers;

public class DiscordDispatcher : IChannelDispatcher
{
    private readonly HttpClient _httpClient;
    private readonly string _webhookUrl;

    public DiscordDispatcher(HttpClient httpClient, string webhookUrl)
    {
        _httpClient = httpClient;
        _webhookUrl = webhookUrl;
        Name = "Discord";
    }

    public string Name { get; }

    public async Task<(bool Success, string? Error)> SendAsync(NotificationRequest request)
    {
        try
        {
            var embed = new
            {
                title = request.Title,
                description = request.Body,
                color = GetColorForPriority(request.Priority),
                fields = new[]
                {
                    new { name = "Priority", value = request.Priority, inline = true },
                    new { name = "Service", value = request.Service ?? "N/A", inline = true }
                },
                timestamp = DateTime.UtcNow.ToString("o")
            };

            var payload = new
            {
                embeds = new[] { embed }
            };

            var json = JsonSerializer.Serialize(payload);
            var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await _httpClient.PostAsync(_webhookUrl, content);
            
            if (response.IsSuccessStatusCode)
            {
                return (true, null);
            }

            var errorContent = await response.Content.ReadAsStringAsync();
            return (false, $"Discord API returned {response.StatusCode}: {errorContent}");
        }
        catch (Exception ex)
        {
            return (false, $"Failed to send Discord notification: {ex.Message}");
        }
    }

    private static int GetColorForPriority(string priority)
    {
        return priority.ToLower() switch
        {
            "high" => 0xFF0000,    // Red
            "normal" => 0x00FF00,  // Green
            "low" => 0xFFFF00,     // Yellow
            _ => 0x808080          // Gray
        };
    }
}
