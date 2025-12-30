using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Endpoints;

public static class ChannelEndpoints
{
    public static void MapChannelEndpoints(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/api/channels").WithTags("Channels");

        // Get all channels
        group.MapGet("/", async (IChannelService channelService) =>
        {
            var channels = await channelService.GetAllChannelsAsync();
            return Results.Ok(channels);
        })
        .WithName("GetChannels")
        .Produces<IEnumerable<ChannelDto>>(200);

        // Get channel by ID
        group.MapGet("/{id:int}", async (int id, IChannelService channelService) =>
        {
            var channel = await channelService.GetChannelByIdAsync(id);
            return channel == null ? Results.NotFound() : Results.Ok(channel);
        })
        .WithName("GetChannel")
        .Produces<ChannelDto>(200)
        .Produces(404);

        // Create channel
        group.MapPost("/", async (CreateChannelRequest request, IChannelService channelService) =>
        {
            var channel = await channelService.CreateChannelAsync(request);
            return Results.Created($"/api/channels/{channel.Id}", channel);
        })
        .WithName("CreateChannel")
        .Produces<ChannelDto>(201);

        // Update channel
        group.MapPut("/{id:int}", async (int id, UpdateChannelRequest request, IChannelService channelService) =>
        {
            var channel = await channelService.UpdateChannelAsync(id, request);
            return channel == null ? Results.NotFound() : Results.Ok(channel);
        })
        .WithName("UpdateChannel")
        .Produces<ChannelDto>(200)
        .Produces(404);

        // Delete channel
        group.MapDelete("/{id:int}", async (int id, IChannelService channelService) =>
        {
            var deleted = await channelService.DeleteChannelAsync(id);
            return deleted ? Results.NoContent() : Results.NotFound();
        })
        .WithName("DeleteChannel")
        .Produces(204)
        .Produces(404);

        // Test channel
        group.MapPost("/{id:int}/test", async (int id, TestChannelRequest request, IChannelService channelService) =>
        {
            var result = await channelService.TestChannelAsync(id, request);
            if (!result.Success)
            {
                return Results.BadRequest(new { success = false, error = result.Error });
            }
            return Results.Ok(new { success = true });
        })
        .WithName("TestChannel")
        .Produces(200)
        .Produces(400);
    }
}
