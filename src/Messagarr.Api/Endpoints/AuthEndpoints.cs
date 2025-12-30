using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;
using Microsoft.AspNetCore.Mvc;

namespace Messagarr.Api.Endpoints;

public static class AuthEndpoints
{
    public static void MapAuthEndpoints(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/api/auth").WithTags("Authentication");

        // Check if setup is complete
        group.MapGet("/setup/status", async (IAuthService authService) =>
        {
            var isComplete = await authService.IsSetupCompleteAsync();
            return Results.Ok(new { setupComplete = isComplete });
        })
        .WithName("GetSetupStatus")
        .Produces(200);

        // Initial setup endpoint
        group.MapPost("/setup", async (SetupRequest request, IAuthService authService) =>
        {
            var result = await authService.SetupInitialUserAsync(request);
            if (!result.Success)
            {
                return Results.BadRequest(result);
            }
            return Results.Ok(result);
        })
        .WithName("Setup")
        .Produces<SetupResponse>(200)
        .Produces<SetupResponse>(400);

        // Login endpoint
        group.MapPost("/login", async (LoginRequest request, IAuthService authService) =>
        {
            var result = await authService.LoginAsync(request);
            if (!result.Success)
            {
                return Results.Unauthorized();
            }
            return Results.Ok(result);
        })
        .WithName("Login")
        .Produces<LoginResponse>(200)
        .Produces(401);
    }
}
