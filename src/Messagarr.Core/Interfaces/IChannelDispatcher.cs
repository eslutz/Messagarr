using Messagarr.Core.Models;

namespace Messagarr.Core.Interfaces;

public interface IChannelDispatcher
{
    string Name { get; }
    Task<(bool Success, string? Error)> SendAsync(NotificationRequest request);
}
