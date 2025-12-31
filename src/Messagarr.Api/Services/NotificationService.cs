using System.Diagnostics;
using System.Text.Json;
using Messagarr.Api.Dispatchers;
using Messagarr.Core.Entities;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;
using Messagarr.Data;
using Microsoft.EntityFrameworkCore;

namespace Messagarr.Api.Services;

public class NotificationService : INotificationService
{
    private readonly MessagearrDbContext _dbContext;
    private readonly ChannelDispatcherFactory _dispatcherFactory;
    private readonly DeduplicationService _deduplicationService;
    private readonly RateLimitService _rateLimitService;
    private readonly RetryService _retryService;

    public NotificationService(
        MessagearrDbContext dbContext,
        ChannelDispatcherFactory dispatcherFactory,
        DeduplicationService deduplicationService,
        RateLimitService rateLimitService,
        RetryService retryService)
    {
        _dbContext = dbContext;
        _dispatcherFactory = dispatcherFactory;
        _deduplicationService = deduplicationService;
        _rateLimitService = rateLimitService;
        _retryService = retryService;
    }

    public async Task<NotificationResponse> SendNotificationAsync(NotificationRequest request)
    {
        var stopwatch = Stopwatch.StartNew();

        // Check for duplicates
        if (_deduplicationService.IsDuplicate(request.Title, request.Body, request.Priority))
        {
            return new NotificationResponse
            {
                Success = false,
                Message = "Duplicate notification detected",
                Duration = $"{stopwatch.ElapsedMilliseconds}ms"
            };
        }

        // Get channels to send to
        var channels = await GetChannelsForNotificationAsync(request);

        if (!channels.Any())
        {
            return new NotificationResponse
            {
                Success = false,
                Message = "No enabled channels found for this priority",
                Duration = $"{stopwatch.ElapsedMilliseconds}ms"
            };
        }

        // Send to all channels in parallel
        var results = new Dictionary<string, ChannelResult>();
        var tasks = channels.Select(async channel =>
        {
            var channelStopwatch = Stopwatch.StartNew();
            var channelResult = await SendToChannelAsync(channel, request);
            channelStopwatch.Stop();

            return new
            {
                ChannelName = channel.Name,
                Result = channelResult,
                DurationMs = (int)channelStopwatch.ElapsedMilliseconds
            };
        });

        var channelResults = await Task.WhenAll(tasks);

        // Create notification history
        var notificationHistory = new NotificationHistory
        {
            Title = request.Title,
            Body = request.Body,
            Priority = request.Priority,
            Service = request.Service,
            EventType = request.EventType,
            Metadata = request.Metadata != null ? JsonSerializer.Serialize(request.Metadata) : null,
            CreatedAt = DateTime.UtcNow,
            Success = channelResults.All(r => r.Result.Success),
            DurationMs = (int)stopwatch.ElapsedMilliseconds
        };

        foreach (var channelResult in channelResults)
        {
            var channel = channels.First(c => c.Name == channelResult.ChannelName);
            
            notificationHistory.ChannelResults.Add(new NotificationChannelResult
            {
                ChannelId = channel.Id,
                ChannelName = channel.Name,
                Success = channelResult.Result.Success,
                ErrorMessage = channelResult.Result.Error,
                DurationMs = channelResult.DurationMs
            });

            results[channelResult.ChannelName] = channelResult.Result;
        }

        _dbContext.NotificationHistories.Add(notificationHistory);
        await _dbContext.SaveChangesAsync();

        stopwatch.Stop();

        return new NotificationResponse
        {
            Success = notificationHistory.Success,
            Message = notificationHistory.Success ? "Notification sent successfully" : "Some channels failed",
            Results = results,
            Duration = $"{stopwatch.ElapsedMilliseconds}ms"
        };
    }

    private async Task<List<Channel>> GetChannelsForNotificationAsync(NotificationRequest request)
    {
        // If specific channels are requested, use those
        if (request.Channels != null && request.Channels.Any())
        {
            return await _dbContext.Channels
                .Where(c => request.Channels.Contains(c.Name) && c.IsEnabled)
                .ToListAsync();
        }

        // Otherwise, use priority-based routing
        // For now, we'll return all enabled channels for any priority
        // This can be enhanced with priority groups configuration
        return await _dbContext.Channels
            .Where(c => c.IsEnabled)
            .ToListAsync();
    }

    private async Task<ChannelResult> SendToChannelAsync(Channel channel, NotificationRequest request)
    {
        try
        {
            // Check rate limit
            if (!await _rateLimitService.TryAcquireAsync(channel.Id.ToString()))
            {
                return new ChannelResult
                {
                    Success = false,
                    Error = "Rate limit exceeded"
                };
            }

            // Create dispatcher
            var dispatcher = _dispatcherFactory.CreateDispatcher(channel);

            // Send with retry
            var result = await _retryService.ExecuteWithRetryAsync(async () =>
            {
                return await dispatcher.SendAsync(request);
            });

            return new ChannelResult
            {
                Success = result.Success,
                Error = result.Error
            };
        }
        catch (Exception ex)
        {
            return new ChannelResult
            {
                Success = false,
                Error = ex.Message
            };
        }
    }
}
