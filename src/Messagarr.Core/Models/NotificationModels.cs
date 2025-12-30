namespace Messagarr.Core.Models;

public class NotificationRequest
{
    public string Title { get; set; } = string.Empty;
    public string Body { get; set; } = string.Empty;
    public string Priority { get; set; } = "normal"; // high, normal, low
    public string? Service { get; set; }
    public string? EventType { get; set; }
    public Dictionary<string, object>? Metadata { get; set; }
    public List<string>? Channels { get; set; } // Optional channel override
}

public class NotificationResponse
{
    public bool Success { get; set; }
    public string Message { get; set; } = string.Empty;
    public Dictionary<string, ChannelResult> Results { get; set; } = new();
    public string Duration { get; set; } = string.Empty;
}

public class ChannelResult
{
    public bool Success { get; set; }
    public string? Error { get; set; }
}
