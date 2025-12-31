using System.Text.Json;
using Messagarr.Api.Dispatchers;
using Messagarr.Core.Entities;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;
using Messagarr.Data;
using Microsoft.EntityFrameworkCore;

namespace Messagarr.Api.Services;

public class ChannelService : IChannelService
{
    private readonly MessagearrDbContext _dbContext;
    private readonly ChannelDispatcherFactory _dispatcherFactory;

    public ChannelService(MessagearrDbContext dbContext, ChannelDispatcherFactory dispatcherFactory)
    {
        _dbContext = dbContext;
        _dispatcherFactory = dispatcherFactory;
    }

    public async Task<IEnumerable<ChannelDto>> GetAllChannelsAsync()
    {
        var channels = await _dbContext.Channels.ToListAsync();
        return channels.Select(MapToDto);
    }

    public async Task<ChannelDto?> GetChannelByIdAsync(int id)
    {
        var channel = await _dbContext.Channels.FindAsync(id);
        return channel == null ? null : MapToDto(channel);
    }

    public async Task<ChannelDto> CreateChannelAsync(CreateChannelRequest request)
    {
        var channelType = ParseChannelType(request.Type);
        
        var channel = new Channel
        {
            Name = request.Name,
            Type = channelType,
            Configuration = JsonSerializer.Serialize(request.Configuration),
            IsEnabled = true,
            CreatedAt = DateTime.UtcNow
        };

        _dbContext.Channels.Add(channel);
        await _dbContext.SaveChangesAsync();

        return MapToDto(channel);
    }

    public async Task<ChannelDto?> UpdateChannelAsync(int id, UpdateChannelRequest request)
    {
        var channel = await _dbContext.Channels.FindAsync(id);
        if (channel == null)
        {
            return null;
        }

        if (!string.IsNullOrEmpty(request.Name))
        {
            channel.Name = request.Name;
        }

        if (request.IsEnabled.HasValue)
        {
            channel.IsEnabled = request.IsEnabled.Value;
        }

        if (request.Configuration != null)
        {
            channel.Configuration = JsonSerializer.Serialize(request.Configuration);
        }

        channel.UpdatedAt = DateTime.UtcNow;
        await _dbContext.SaveChangesAsync();

        return MapToDto(channel);
    }

    public async Task<bool> DeleteChannelAsync(int id)
    {
        var channel = await _dbContext.Channels.FindAsync(id);
        if (channel == null)
        {
            return false;
        }

        _dbContext.Channels.Remove(channel);
        await _dbContext.SaveChangesAsync();

        return true;
    }

    public async Task<(bool Success, string? Error)> TestChannelAsync(int id, TestChannelRequest request)
    {
        var channel = await _dbContext.Channels.FindAsync(id);
        if (channel == null)
        {
            return (false, "Channel not found");
        }

        if (!channel.IsEnabled)
        {
            return (false, "Channel is disabled");
        }

        // Create a test notification request
        var notificationRequest = new NotificationRequest
        {
            Title = request.Title,
            Body = request.Body,
            Priority = "normal",
            Service = "Messagarr",
            EventType = "test"
        };

        // Get the appropriate dispatcher and send the test notification
        var dispatcher = GetDispatcherForChannel(channel);
        if (dispatcher == null)
        {
            return (false, $"No dispatcher found for channel type {channel.Type}");
        }

        var result = await dispatcher.SendAsync(notificationRequest);

        // Update test statistics
        channel.LastTestedAt = DateTime.UtcNow;
        if (result.Success)
        {
            channel.TestSuccessCount = (channel.TestSuccessCount ?? 0) + 1;
        }
        else
        {
            channel.TestFailureCount = (channel.TestFailureCount ?? 0) + 1;
        }
        await _dbContext.SaveChangesAsync();

        return result;
    }

    private IChannelDispatcher? GetDispatcherForChannel(Channel channel)
    {
        try
        {
            return _dispatcherFactory.CreateDispatcher(channel);
        }
        catch (Exception)
        {
            return null;
        }
    }

    private static ChannelDto MapToDto(Channel channel)
    {
        return new ChannelDto
        {
            Id = channel.Id,
            Name = channel.Name,
            Type = channel.Type.ToString().ToLower(),
            IsEnabled = channel.IsEnabled,
            CreatedAt = channel.CreatedAt,
            UpdatedAt = channel.UpdatedAt,
            TestSuccessCount = channel.TestSuccessCount,
            TestFailureCount = channel.TestFailureCount,
            LastTestedAt = channel.LastTestedAt
        };
    }

    private static ChannelType ParseChannelType(string type)
    {
        return type.ToLower() switch
        {
            "email" => ChannelType.Email,
            "discord" => ChannelType.Discord,
            "slack" => ChannelType.Slack,
            "teams" => ChannelType.Teams,
            _ => throw new ArgumentException($"Unknown channel type: {type}")
        };
    }
}
