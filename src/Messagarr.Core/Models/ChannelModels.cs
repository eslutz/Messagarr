using Messagarr.Core.Entities;

namespace Messagarr.Core.Models;

public class ChannelDto
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public string Type { get; set; } = string.Empty;
    public bool IsEnabled { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime? UpdatedAt { get; set; }
    public int? TestSuccessCount { get; set; }
    public int? TestFailureCount { get; set; }
    public DateTime? LastTestedAt { get; set; }
}

public class CreateChannelRequest
{
    public string Name { get; set; } = string.Empty;
    public string Type { get; set; } = string.Empty; // email, discord, slack, teams
    public ChannelConfiguration Configuration { get; set; } = new();
}

public class UpdateChannelRequest
{
    public string? Name { get; set; }
    public bool? IsEnabled { get; set; }
    public ChannelConfiguration? Configuration { get; set; }
}

public class ChannelConfiguration
{
    // Email specific
    public string? SmtpHost { get; set; }
    public int? SmtpPort { get; set; }
    public string? SmtpUser { get; set; }
    public string? SmtpPassword { get; set; }
    public string? FromEmail { get; set; }
    public string? ToEmail { get; set; }
    
    // Webhook specific (Discord, Slack, Teams)
    public string? WebhookUrl { get; set; }
}

public class TestChannelRequest
{
    public string Title { get; set; } = "Test Notification";
    public string Body { get; set; } = "This is a test notification from Messagarr";
}
