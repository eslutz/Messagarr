using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Endpoints;

public static class NotificationEndpoints
{
    public static void MapNotificationEndpoints(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/api/notify").WithTags("Notifications");

        // Send notification
        group.MapPost("/", async (NotificationRequest request, INotificationService notificationService) =>
        {
            var result = await notificationService.SendNotificationAsync(request);
            
            if (!result.Success)
            {
                return Results.BadRequest(result);
            }
            
            return Results.Ok(result);
        })
        .WithName("SendNotification")
        .Produces<NotificationResponse>(200)
        .Produces<NotificationResponse>(400);
    }
}
