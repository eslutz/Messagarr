namespace Messagarr.Core.Entities;

public class Channel
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public ChannelType Type { get; set; }
    public string Configuration { get; set; } = string.Empty; // JSON configuration
    public bool IsEnabled { get; set; } = true;
    public DateTime CreatedAt { get; set; }
    public DateTime? UpdatedAt { get; set; }
    public int? TestSuccessCount { get; set; } = 0;
    public int? TestFailureCount { get; set; } = 0;
    public DateTime? LastTestedAt { get; set; }
}

public enum ChannelType
{
    Email = 1,
    Discord = 2,
    Slack = 3,
    Teams = 4
}
