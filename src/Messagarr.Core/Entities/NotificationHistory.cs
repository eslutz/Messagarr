namespace Messagarr.Core.Entities;

public class NotificationHistory
{
    public int Id { get; set; }
    public string Title { get; set; } = string.Empty;
    public string Body { get; set; } = string.Empty;
    public string Priority { get; set; } = "normal";
    public string? Service { get; set; }
    public string? EventType { get; set; }
    public string? Metadata { get; set; } // JSON metadata
    public DateTime CreatedAt { get; set; }
    public bool Success { get; set; }
    public string? ErrorMessage { get; set; }
    public int DurationMs { get; set; }
    
    public ICollection<NotificationChannelResult> ChannelResults { get; set; } = new List<NotificationChannelResult>();
}

public class NotificationChannelResult
{
    public int Id { get; set; }
    public int NotificationHistoryId { get; set; }
    public NotificationHistory NotificationHistory { get; set; } = null!;
    public int ChannelId { get; set; }
    public string ChannelName { get; set; } = string.Empty;
    public bool Success { get; set; }
    public string? ErrorMessage { get; set; }
    public int DurationMs { get; set; }
}
