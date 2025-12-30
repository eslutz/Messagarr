using Messagarr.Core.Models;

namespace Messagarr.Core.Interfaces;

public interface INotificationService
{
    Task<NotificationResponse> SendNotificationAsync(NotificationRequest request);
}
