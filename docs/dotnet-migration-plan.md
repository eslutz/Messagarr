# Messagarr .NET 10 Migration Plan

> **Scope**: Complete rewrite from Go to C#/.NET 10 with minimal APIs, SQLite persistence, and *arr-style web UI

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Solution Structure](#2-solution-structure)
3. [NuGet Packages](#3-nuget-packages)
4. [Domain Models & EF Core Entities](#4-domain-models--ef-core-entities)
5. [Service Layer Architecture](#5-service-layer-architecture)
6. [Minimal API Endpoints](#6-minimal-api-endpoints)
7. [Authentication System](#7-authentication-system)
8. [Web UI Recommendation](#8-web-ui-recommendation)
9. [Migration Path](#9-migration-path)
10. [Docker & Deployment](#10-docker--deployment)
11. [Development Workflow](#11-development-workflow)
12. [Implementation Phases](#12-implementation-phases)

---

## 1. Executive Summary

### Current Go Architecture Mapping

| Go Component | .NET Equivalent |
|--------------|-----------------|
| `cmd/messagarr/main.go` | `Program.cs` (minimal API host) |
| `internal/api/server.go` | Minimal API endpoint groups |
| `internal/channels/` | `Services/Channels/` with `INotificationChannel` interface |
| `internal/config/config.go` | `appsettings.json` + EF Core SQLite |
| `internal/resilience/` | Polly library for retry/circuit breaker |
| `internal/models/` | `Models/` DTOs + EF Core entities |
| YAML config | SQLite database (runtime configurable via UI) |

### Key Differences from Go Version

1. **Configuration moves to database** - Channels, priority groups, and settings stored in SQLite (like *arr apps)
2. **First-launch wizard** - Creates admin user and API key on initial startup
3. **Web UI** - Full management interface for channels, users, and notification history
4. **Auth required** - API key for programmatic access, cookie auth for UI

---

## 2. Solution Structure

```
Messagarr/
├── Messagarr.sln
├── src/
│   └── Messagarr/
│       ├── Messagarr.csproj
│       ├── Program.cs                          # Entry point, DI, middleware
│       ├── appsettings.json                    # Static config (ports, paths)
│       ├── appsettings.Development.json
│       │
│       ├── Data/
│       │   ├── MessagarrDbContext.cs           # EF Core DbContext
│       │   ├── Migrations/                     # EF Core migrations
│       │   └── Seeding/
│       │       └── DatabaseSeeder.cs           # First-launch setup
│       │
│       ├── Entities/                           # EF Core entities (DB tables)
│       │   ├── Channel.cs
│       │   ├── PriorityGroup.cs
│       │   ├── User.cs
│       │   ├── ApiKey.cs
│       │   ├── NotificationLog.cs
│       │   └── Setting.cs
│       │
│       ├── Models/                             # DTOs (API contracts)
│       │   ├── Requests/
│       │   │   ├── NotificationRequest.cs
│       │   │   ├── ChannelRequest.cs
│       │   │   └── LoginRequest.cs
│       │   ├── Responses/
│       │   │   ├── NotificationResponse.cs
│       │   │   ├── HealthResponse.cs
│       │   │   └── StatusResponse.cs
│       │   └── Mapping/
│       │       └── MappingExtensions.cs        # Entity <-> DTO mapping
│       │
│       ├── Services/
│       │   ├── Channels/
│       │   │   ├── INotificationChannel.cs     # Channel interface
│       │   │   ├── ChannelFactory.cs           # Creates channel instances
│       │   │   ├── DiscordChannel.cs
│       │   │   ├── SlackChannel.cs
│       │   │   ├── TeamsChannel.cs
│       │   │   └── SmtpChannel.cs
│       │   ├── Dispatch/
│       │   │   ├── INotificationDispatcher.cs
│       │   │   └── NotificationDispatcher.cs   # Parallel dispatch + routing
│       │   ├── Resilience/
│       │   │   ├── Deduplicator.cs              # SHA256 hash + TTL cache
│       │   │   └── ResiliencePolicies.cs        # Polly policies
│       │   ├── Auth/
│       │   │   ├── IAuthService.cs
│       │   │   ├── AuthService.cs
│       │   │   └── ApiKeyAuthHandler.cs
│       │   └── Settings/
│       │       ├── ISettingsService.cs
│       │       └── SettingsService.cs
│       │
│       ├── Endpoints/                          # Minimal API endpoint groups
│       │   ├── NotificationEndpoints.cs
│       │   ├── HealthEndpoints.cs
│       │   ├── ChannelEndpoints.cs
│       │   ├── UserEndpoints.cs
│       │   ├── SettingsEndpoints.cs
│       │   └── AuthEndpoints.cs
│       │
│       ├── Middleware/
│       │   ├── RequestLoggingMiddleware.cs
│       │   └── FirstLaunchMiddleware.cs        # Redirects to wizard if no users
│       │
│       ├── wwwroot/                            # Static files (if using SPA)
│       │   ├── index.html
│       │   └── assets/
│       │
│       └── ClientApp/                          # React SPA source (build output → wwwroot)
│           ├── package.json
│           ├── vite.config.ts
│           └── src/
│               ├── App.tsx
│               ├── api/
│               ├── components/
│               ├── pages/
│               └── styles/
│
├── tests/
│   └── Messagarr.Tests/
│       ├── Messagarr.Tests.csproj
│       ├── Endpoints/
│       ├── Services/
│       └── Integration/
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## 3. NuGet Packages

### Core Packages

```xml
<ItemGroup>
  <!-- ASP.NET Core (included in SDK) -->
  <PackageReference Include="Microsoft.AspNetCore.OpenApi" Version="10.0.0" />
  <PackageReference Include="Swashbuckle.AspNetCore" Version="7.0.0" />

  <!-- EF Core + SQLite -->
  <PackageReference Include="Microsoft.EntityFrameworkCore.Sqlite" Version="10.0.0" />
  <PackageReference Include="Microsoft.EntityFrameworkCore.Design" Version="10.0.0" />

  <!-- Auth -->
  <PackageReference Include="Microsoft.AspNetCore.Authentication.JwtBearer" Version="10.0.0" />

  <!-- Resilience (replaces Go retry/rate limit) -->
  <PackageReference Include="Microsoft.Extensions.Http.Polly" Version="10.0.0" />
  <PackageReference Include="Polly" Version="8.5.0" />

  <!-- Metrics -->
  <PackageReference Include="prometheus-net.AspNetCore" Version="8.2.1" />

  <!-- Utilities -->
  <PackageReference Include="Serilog.AspNetCore" Version="9.0.0" />
  <PackageReference Include="MailKit" Version="4.8.0" />

  <!-- Caching (for deduplication) -->
  <PackageReference Include="Microsoft.Extensions.Caching.Memory" Version="10.0.0" />
</ItemGroup>
```

### Test Packages

```xml
<ItemGroup>
  <PackageReference Include="Microsoft.AspNetCore.Mvc.Testing" Version="10.0.0" />
  <PackageReference Include="xunit" Version="2.9.0" />
  <PackageReference Include="Moq" Version="4.20.0" />
  <PackageReference Include="FluentAssertions" Version="7.0.0" />
</ItemGroup>
```

---

## 4. Domain Models & EF Core Entities

### 4.1 EF Core Entities

```csharp
// Entities/Channel.cs
public class Channel
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public ChannelType Type { get; set; }
    public string? WebhookUrl { get; set; }
    public string? SmtpHost { get; set; }
    public int? SmtpPort { get; set; }
    public string? SmtpUser { get; set; }
    public string? SmtpPassword { get; set; }  // Encrypted at rest
    public string? SmtpFrom { get; set; }
    public string? SmtpTo { get; set; }
    public bool IsEnabled { get; set; } = true;
    public DateTime CreatedAt { get; set; }
    public DateTime? UpdatedAt { get; set; }

    // Navigation
    public ICollection<PriorityGroupChannel> PriorityGroups { get; set; } = [];
}

public enum ChannelType
{
    Discord,
    Slack,
    Teams,
    Smtp
}

// Entities/PriorityGroup.cs
public class PriorityGroup
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;  // "high", "normal", "low"
    public int SortOrder { get; set; }

    // Many-to-many with Channel
    public ICollection<PriorityGroupChannel> Channels { get; set; } = [];
}

// Entities/PriorityGroupChannel.cs (join table)
public class PriorityGroupChannel
{
    public int PriorityGroupId { get; set; }
    public PriorityGroup PriorityGroup { get; set; } = null!;

    public int ChannelId { get; set; }
    public Channel Channel { get; set; } = null!;
}

// Entities/User.cs
public class User
{
    public int Id { get; set; }
    public string Username { get; set; } = string.Empty;
    public string PasswordHash { get; set; } = string.Empty;  // BCrypt
    public bool IsAdmin { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime? LastLoginAt { get; set; }

    public ICollection<ApiKey> ApiKeys { get; set; } = [];
}

// Entities/ApiKey.cs
public class ApiKey
{
    public int Id { get; set; }
    public string Key { get; set; } = string.Empty;  // Stored hashed
    public string KeyPrefix { get; set; } = string.Empty;  // First 8 chars for display
    public string? Description { get; set; }
    public int UserId { get; set; }
    public User User { get; set; } = null!;
    public DateTime CreatedAt { get; set; }
    public DateTime? ExpiresAt { get; set; }
    public DateTime? LastUsedAt { get; set; }
}

// Entities/NotificationLog.cs
public class NotificationLog
{
    public long Id { get; set; }
    public string Title { get; set; } = string.Empty;
    public string? Body { get; set; }
    public string? Priority { get; set; }
    public string? Service { get; set; }
    public string? EventType { get; set; }
    public string ResultsJson { get; set; } = "{}";  // Serialized channel results
    public bool Success { get; set; }
    public int DurationMs { get; set; }
    public DateTime CreatedAt { get; set; }
    public string? DeduplicationHash { get; set; }
}

// Entities/Setting.cs
public class Setting
{
    public string Key { get; set; } = string.Empty;
    public string Value { get; set; } = string.Empty;
    public string? Description { get; set; }
}
```

### 4.2 DbContext

```csharp
// Data/MessagarrDbContext.cs
public class MessagarrDbContext : DbContext
{
    public DbSet<Channel> Channels => Set<Channel>();
    public DbSet<PriorityGroup> PriorityGroups => Set<PriorityGroup>();
    public DbSet<PriorityGroupChannel> PriorityGroupChannels => Set<PriorityGroupChannel>();
    public DbSet<User> Users => Set<User>();
    public DbSet<ApiKey> ApiKeys => Set<ApiKey>();
    public DbSet<NotificationLog> NotificationLogs => Set<NotificationLog>();
    public DbSet<Setting> Settings => Set<Setting>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        // Composite key for join table
        modelBuilder.Entity<PriorityGroupChannel>()
            .HasKey(pgc => new { pgc.PriorityGroupId, pgc.ChannelId });

        // Setting key is the primary key
        modelBuilder.Entity<Setting>()
            .HasKey(s => s.Key);

        // Indexes
        modelBuilder.Entity<NotificationLog>()
            .HasIndex(n => n.CreatedAt);

        modelBuilder.Entity<NotificationLog>()
            .HasIndex(n => n.DeduplicationHash);

        modelBuilder.Entity<ApiKey>()
            .HasIndex(a => a.Key);
    }
}
```

### 4.3 Request/Response DTOs

```csharp
// Models/Requests/NotificationRequest.cs
public record NotificationRequest(
    string? Title,
    string? Body,
    string? Priority,
    string? Service,
    string? EventType,
    Dictionary<string, string>? Metadata,
    List<string>? Channels
);

// Models/Responses/NotificationResponse.cs
public record NotificationResponse(
    Dictionary<string, ChannelResult> Results,
    string Message,
    string Duration,
    bool Success
);

public record ChannelResult(bool Success, string? Error);

// Models/Responses/HealthResponse.cs
public record HealthResponse(string Status, DateTime Timestamp, string Version);

// Models/Responses/StatusResponse.cs
public record StatusResponse(
    string Uptime,
    long TotalNotifications,
    long FailedNotifications,
    Dictionary<string, long> ChannelStats
);
```

---

## 5. Service Layer Architecture

### 5.1 Channel Interface & Implementations

```csharp
// Services/Channels/INotificationChannel.cs
public interface INotificationChannel
{
    string Name { get; }
    ChannelType Type { get; }
    Task<Result> SendAsync(NotificationRequest request, CancellationToken ct = default);
}

// Result type for channel operations
public record Result(bool Success, string? Error = null);

// Services/Channels/DiscordChannel.cs
public class DiscordChannel : INotificationChannel
{
    private readonly HttpClient _httpClient;
    private readonly Channel _config;

    public string Name => _config.Name;
    public ChannelType Type => ChannelType.Discord;

    public DiscordChannel(Channel config, HttpClient httpClient)
    {
        _config = config;
        _httpClient = httpClient;
    }

    public async Task<Result> SendAsync(NotificationRequest request, CancellationToken ct)
    {
        var embed = new DiscordEmbed
        {
            Title = request.Title,
            Description = request.Body,
            Color = GetColorForPriority(request.Priority)
        };

        // Add metadata fields...

        var webhook = new { embeds = new[] { embed } };
        var response = await _httpClient.PostAsJsonAsync(_config.WebhookUrl, webhook, ct);

        return response.IsSuccessStatusCode
            ? new Result(true)
            : new Result(false, $"HTTP {response.StatusCode}");
    }
}
```

### 5.2 Channel Factory

```csharp
// Services/Channels/ChannelFactory.cs
public class ChannelFactory
{
    private readonly IHttpClientFactory _httpClientFactory;

    public INotificationChannel Create(Channel config)
    {
        var httpClient = _httpClientFactory.CreateClient("Notifications");

        return config.Type switch
        {
            ChannelType.Discord => new DiscordChannel(config, httpClient),
            ChannelType.Slack => new SlackChannel(config, httpClient),
            ChannelType.Teams => new TeamsChannel(config, httpClient),
            ChannelType.Smtp => new SmtpChannel(config),
            _ => throw new NotSupportedException($"Channel type {config.Type} not supported")
        };
    }
}
```

### 5.3 Notification Dispatcher

```csharp
// Services/Dispatch/NotificationDispatcher.cs
public class NotificationDispatcher : INotificationDispatcher
{
    private readonly MessagarrDbContext _db;
    private readonly ChannelFactory _channelFactory;
    private readonly Deduplicator _deduplicator;
    private readonly ResiliencePipeline _pipeline;
    private readonly ILogger<NotificationDispatcher> _logger;

    public async Task<NotificationResponse> DispatchAsync(
        NotificationRequest request,
        CancellationToken ct)
    {
        var stopwatch = Stopwatch.StartNew();

        // Check deduplication
        var hash = _deduplicator.ComputeHash(request);
        if (_deduplicator.IsDuplicate(hash))
        {
            return new NotificationResponse(
                Results: new Dictionary<string, ChannelResult>(),
                Message: "Duplicate notification suppressed",
                Duration: $"{stopwatch.ElapsedMilliseconds}ms",
                Success: true
            );
        }

        // Resolve target channels
        var channelConfigs = await ResolveChannelsAsync(request, ct);

        // Dispatch in parallel with resilience
        var tasks = channelConfigs.Select(async config =>
        {
            var channel = _channelFactory.Create(config);
            var result = await _pipeline.ExecuteAsync(
                async token => await channel.SendAsync(request, token), ct);
            return (config.Name, result);
        });

        var results = await Task.WhenAll(tasks);

        // Build response
        var resultDict = results.ToDictionary(
            r => r.Name,
            r => new ChannelResult(r.result.Success, r.result.Error));

        var allSucceeded = results.All(r => r.result.Success);

        // Log to database
        await LogNotificationAsync(request, resultDict, stopwatch.ElapsedMilliseconds, allSucceeded, hash, ct);

        return new NotificationResponse(
            Results: resultDict,
            Message: allSucceeded ? "Notification sent" : "Some channels failed",
            Duration: $"{stopwatch.ElapsedMilliseconds}ms",
            Success: allSucceeded
        );
    }

    private async Task<List<Channel>> ResolveChannelsAsync(NotificationRequest request, CancellationToken ct)
    {
        // If explicit channels specified, use those
        if (request.Channels?.Count > 0)
        {
            return await _db.Channels
                .Where(c => request.Channels.Contains(c.Name) && c.IsEnabled)
                .ToListAsync(ct);
        }

        // Otherwise, route by priority group
        var priority = request.Priority ?? "normal";
        return await _db.PriorityGroups
            .Where(pg => pg.Name == priority)
            .SelectMany(pg => pg.Channels.Select(pgc => pgc.Channel))
            .Where(c => c.IsEnabled)
            .ToListAsync(ct);
    }
}
```

### 5.4 Resilience with Polly

```csharp
// Services/Resilience/ResiliencePolicies.cs
public static class ResiliencePolicies
{
    public static ResiliencePipeline<Result> CreateNotificationPipeline()
    {
        return new ResiliencePipelineBuilder<Result>()
            // Retry with exponential backoff (like Go: 1s, 2s, 4s, max 30s)
            .AddRetry(new RetryStrategyOptions<Result>
            {
                MaxRetryAttempts = 3,
                Delay = TimeSpan.FromSeconds(1),
                BackoffType = DelayBackoffType.Exponential,
                MaxDelay = TimeSpan.FromSeconds(30),
                ShouldHandle = new PredicateBuilder<Result>()
                    .HandleResult(r => !r.Success)
            })
            // Rate limiting (like Go: 10 req/sec per channel)
            .AddRateLimiter(new SlidingWindowRateLimiter(new SlidingWindowRateLimiterOptions
            {
                PermitLimit = 10,
                Window = TimeSpan.FromSeconds(1),
                SegmentsPerWindow = 2
            }))
            .Build();
    }
}
```

### 5.5 Deduplicator

```csharp
// Services/Resilience/Deduplicator.cs
public class Deduplicator
{
    private readonly IMemoryCache _cache;
    private readonly TimeSpan _ttl;

    public Deduplicator(IMemoryCache cache, IOptions<DeduplicationOptions> options)
    {
        _cache = cache;
        _ttl = options.Value.Ttl;
    }

    public string ComputeHash(NotificationRequest request)
    {
        var data = JsonSerializer.Serialize(new
        {
            request.Title,
            request.Body,
            request.Priority,
            request.Service,
            request.EventType
        });

        var hashBytes = SHA256.HashData(Encoding.UTF8.GetBytes(data));
        return Convert.ToHexString(hashBytes);
    }

    public bool IsDuplicate(string hash)
    {
        if (_cache.TryGetValue(hash, out _))
            return true;

        _cache.Set(hash, true, _ttl);
        return false;
    }
}
```

---

## 6. Minimal API Endpoints

### 6.1 Program.cs Setup

```csharp
// Program.cs
var builder = WebApplication.CreateBuilder(args);

// Add services
builder.Services.AddDbContext<MessagarrDbContext>(options =>
    options.UseSqlite(builder.Configuration.GetConnectionString("Default")));

builder.Services.AddMemoryCache();
builder.Services.AddHttpClient("Notifications");

// Custom services
builder.Services.AddScoped<INotificationDispatcher, NotificationDispatcher>();
builder.Services.AddScoped<IAuthService, AuthService>();
builder.Services.AddSingleton<ChannelFactory>();
builder.Services.AddSingleton<Deduplicator>();
builder.Services.AddSingleton(ResiliencePolicies.CreateNotificationPipeline());

// Auth
builder.Services.AddAuthentication()
    .AddScheme<ApiKeyAuthOptions, ApiKeyAuthHandler>("ApiKey", null)
    .AddCookie("Cookie", options =>
    {
        options.LoginPath = "/login";
        options.Cookie.Name = "Messagarr.Auth";
    });

builder.Services.AddAuthorization(options =>
{
    options.AddPolicy("ApiAccess", policy =>
        policy.RequireAuthenticatedUser()
              .AddAuthenticationSchemes("ApiKey", "Cookie"));
});

// Swagger
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();

// Prometheus
builder.Services.AddSingleton<NotificationMetrics>();

// Serilog
builder.Host.UseSerilog((ctx, config) =>
    config.ReadFrom.Configuration(ctx.Configuration));

var app = builder.Build();

// Middleware
app.UseSerilogRequestLogging();
app.UseMiddleware<FirstLaunchMiddleware>();

if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseStaticFiles();  // Serves wwwroot (React SPA)
app.UseAuthentication();
app.UseAuthorization();
app.UseHttpMetrics();  // Prometheus

// Map endpoints
app.MapNotificationEndpoints();
app.MapHealthEndpoints();
app.MapChannelEndpoints();
app.MapUserEndpoints();
app.MapAuthEndpoints();
app.MapSettingsEndpoints();
app.MapMetrics();  // /metrics for Prometheus

// SPA fallback
app.MapFallbackToFile("index.html");

// Apply migrations on startup
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<MessagarrDbContext>();
    await db.Database.MigrateAsync();
}

app.Run();
```

### 6.2 Endpoint Groups

```csharp
// Endpoints/NotificationEndpoints.cs
public static class NotificationEndpoints
{
    public static void MapNotificationEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api")
            .RequireAuthorization("ApiAccess")
            .WithTags("Notifications");

        group.MapPost("/notify", SendNotification)
            .WithName("SendNotification")
            .Produces<NotificationResponse>(200)
            .Produces<NotificationResponse>(207)
            .ProducesValidationProblem();
    }

    private static async Task<IResult> SendNotification(
        NotificationRequest request,
        INotificationDispatcher dispatcher,
        NotificationMetrics metrics,
        CancellationToken ct)
    {
        if (string.IsNullOrEmpty(request.Title) && string.IsNullOrEmpty(request.Body))
            return Results.BadRequest("Title or body is required");

        var response = await dispatcher.DispatchAsync(request, ct);

        // Update Prometheus metrics
        metrics.RecordNotification(response);

        var statusCode = response.Success ? 200 : 207;
        return Results.Json(response, statusCode: statusCode);
    }
}

// Endpoints/HealthEndpoints.cs
public static class HealthEndpoints
{
    public static void MapHealthEndpoints(this WebApplication app)
    {
        // Public endpoints (no auth)
        app.MapGet("/health", () => new HealthResponse(
            Status: "ok",
            Timestamp: DateTime.UtcNow,
            Version: Assembly.GetExecutingAssembly().GetName().Version?.ToString() ?? "0.0.0"
        )).WithTags("Health");

        app.MapGet("/ready", async (MessagarrDbContext db) =>
        {
            var checks = new List<ReadyCheck>();

            // Database check
            try
            {
                await db.Database.CanConnectAsync();
                checks.Add(new ReadyCheck("database", "ok", "Connected"));
            }
            catch (Exception ex)
            {
                checks.Add(new ReadyCheck("database", "error", ex.Message));
            }

            var ready = checks.All(c => c.Status == "ok");
            return new ReadyResponse(ready, DateTime.UtcNow, checks);
        }).WithTags("Health");

        app.MapGet("/status", async (
            MessagarrDbContext db,
            IHostApplicationLifetime lifetime) =>
        {
            var stats = await db.NotificationLogs
                .GroupBy(_ => 1)
                .Select(g => new
                {
                    Total = g.Count(),
                    Failed = g.Count(n => !n.Success)
                })
                .FirstOrDefaultAsync();

            return new StatusResponse(
                Uptime: (DateTime.UtcNow - Process.GetCurrentProcess().StartTime.ToUniversalTime()).ToString(),
                TotalNotifications: stats?.Total ?? 0,
                FailedNotifications: stats?.Failed ?? 0,
                ChannelStats: new Dictionary<string, long>()  // Populate from logs
            );
        }).RequireAuthorization("ApiAccess").WithTags("Health");
    }
}

// Endpoints/ChannelEndpoints.cs (UI management)
public static class ChannelEndpoints
{
    public static void MapChannelEndpoints(this WebApplication app)
    {
        var group = app.MapGroup("/api/channels")
            .RequireAuthorization("ApiAccess")
            .WithTags("Channels");

        group.MapGet("/", async (MessagarrDbContext db) =>
            await db.Channels.ToListAsync());

        group.MapGet("/{id:int}", async (int id, MessagarrDbContext db) =>
            await db.Channels.FindAsync(id) is Channel c
                ? Results.Ok(c)
                : Results.NotFound());

        group.MapPost("/", async (ChannelRequest request, MessagarrDbContext db) =>
        {
            var channel = request.ToEntity();
            db.Channels.Add(channel);
            await db.SaveChangesAsync();
            return Results.Created($"/api/channels/{channel.Id}", channel);
        });

        group.MapPut("/{id:int}", async (int id, ChannelRequest request, MessagarrDbContext db) =>
        {
            var channel = await db.Channels.FindAsync(id);
            if (channel is null) return Results.NotFound();

            request.UpdateEntity(channel);
            await db.SaveChangesAsync();
            return Results.Ok(channel);
        });

        group.MapDelete("/{id:int}", async (int id, MessagarrDbContext db) =>
        {
            var channel = await db.Channels.FindAsync(id);
            if (channel is null) return Results.NotFound();

            db.Channels.Remove(channel);
            await db.SaveChangesAsync();
            return Results.NoContent();
        });

        group.MapPost("/{id:int}/test", async (int id, MessagarrDbContext db, ChannelFactory factory) =>
        {
            var channel = await db.Channels.FindAsync(id);
            if (channel is null) return Results.NotFound();

            var notificationChannel = factory.Create(channel);
            var result = await notificationChannel.SendAsync(new NotificationRequest(
                Title: "Test Notification",
                Body: "This is a test from Messagarr",
                Priority: "normal",
                Service: null, EventType: null, Metadata: null, Channels: null
            ));

            return result.Success ? Results.Ok("Test sent") : Results.BadRequest(result.Error);
        });
    }
}
```

---

## 7. Authentication System

### 7.1 Recommendation: Custom Lightweight Auth (not ASP.NET Identity)

**Rationale**: *arr apps use simple custom auth (not Identity). Keep it minimal:

| Feature | Implementation |
|---------|----------------|
| Password storage | BCrypt hash |
| API keys | Random 32-byte + SHA256 hash storage |
| Session | Cookie with encrypted claims |
| First launch | Middleware redirects to `/initialize` if no users exist |

### 7.2 API Key Auth Handler

```csharp
// Services/Auth/ApiKeyAuthHandler.cs
public class ApiKeyAuthHandler : AuthenticationHandler<ApiKeyAuthOptions>
{
    private readonly MessagarrDbContext _db;

    protected override async Task<AuthenticateResult> HandleAuthenticateAsync()
    {
        // Check X-Api-Key header
        if (!Request.Headers.TryGetValue("X-Api-Key", out var headerValue))
            return AuthenticateResult.NoResult();

        var providedKey = headerValue.ToString();
        var keyHash = ComputeKeyHash(providedKey);

        var apiKey = await _db.ApiKeys
            .Include(k => k.User)
            .FirstOrDefaultAsync(k => k.Key == keyHash);

        if (apiKey is null)
            return AuthenticateResult.Fail("Invalid API key");

        if (apiKey.ExpiresAt.HasValue && apiKey.ExpiresAt < DateTime.UtcNow)
            return AuthenticateResult.Fail("API key expired");

        // Update last used
        apiKey.LastUsedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();

        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, apiKey.UserId.ToString()),
            new Claim(ClaimTypes.Name, apiKey.User.Username),
            new Claim("ApiKeyId", apiKey.Id.ToString())
        };

        var identity = new ClaimsIdentity(claims, Scheme.Name);
        var principal = new ClaimsPrincipal(identity);
        var ticket = new AuthenticationTicket(principal, Scheme.Name);

        return AuthenticateResult.Success(ticket);
    }
}
```

### 7.3 First Launch Middleware

```csharp
// Middleware/FirstLaunchMiddleware.cs
public class FirstLaunchMiddleware
{
    private readonly RequestDelegate _next;
    private bool? _isInitialized;

    public async Task InvokeAsync(HttpContext context, MessagarrDbContext db)
    {
        // Skip for static files and initialize endpoint
        var path = context.Request.Path.Value ?? "";
        if (path.StartsWith("/assets") ||
            path.StartsWith("/api/initialize") ||
            path == "/initialize")
        {
            await _next(context);
            return;
        }

        // Cache the check
        _isInitialized ??= await db.Users.AnyAsync();

        if (!_isInitialized.Value)
        {
            // Redirect to initialization wizard
            if (context.Request.Path.StartsWithSegments("/api"))
            {
                context.Response.StatusCode = 503;
                await context.Response.WriteAsJsonAsync(new
                {
                    error = "System not initialized",
                    initializeUrl = "/initialize"
                });
                return;
            }

            context.Response.Redirect("/initialize");
            return;
        }

        await _next(context);
    }
}
```

---

## 8. Web UI Recommendation

### 8.1 Recommendation: **React + Vite SPA**

**Why React over Blazor/Razor Pages:**

| Factor | React | Blazor Server | Razor + htmx |
|--------|-------|---------------|--------------|
| *arr Consistency | ✅ Sonarr/Radarr/etc use React | ❌ Different paradigm | ❌ |
| Developer Pool | ✅ Large | ⚠️ Smaller | ⚠️ Niche |
| Dark Theme | ✅ Easy with CSS-in-JS | ✅ Possible | ✅ Possible |
| Offline Capable | ✅ SPA bundles | ❌ Server connection required | ⚠️ |
| Build Complexity | ⚠️ Separate build | ✅ Integrated | ✅ Integrated |

### 8.2 UI Architecture

```
ClientApp/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── index.html
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── api/
    │   ├── client.ts           # Fetch wrapper with API key
    │   ├── channels.ts
    │   ├── notifications.ts
    │   └── auth.ts
    ├── components/
    │   ├── Layout/
    │   │   ├── Sidebar.tsx     # *arr-style sidebar nav
    │   │   ├── Header.tsx
    │   │   └── Layout.tsx
    │   ├── Channels/
    │   │   ├── ChannelCard.tsx
    │   │   ├── ChannelModal.tsx
    │   │   └── ChannelList.tsx
    │   ├── Common/
    │   │   ├── Card.tsx
    │   │   ├── Modal.tsx
    │   │   ├── Button.tsx
    │   │   └── Input.tsx
    │   └── Forms/
    │       └── ChannelForm.tsx
    ├── pages/
    │   ├── Dashboard.tsx
    │   ├── Channels.tsx
    │   ├── History.tsx
    │   ├── Settings.tsx
    │   ├── Users.tsx
    │   ├── Login.tsx
    │   └── Initialize.tsx      # First-launch wizard
    ├── hooks/
    │   ├── useAuth.ts
    │   └── useApi.ts
    └── styles/
        ├── globals.css         # CSS variables for dark theme
        └── theme.ts
```

### 8.3 Recommended Libraries

```json
{
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router-dom": "^7.0.0",
    "@tanstack/react-query": "^5.0.0",  // Data fetching
    "lucide-react": "^0.460.0",          // Icons
    "clsx": "^2.1.0",                     // Class names
    "zustand": "^5.0.0"                   // State management (simpler than Redux)
  },
  "devDependencies": {
    "vite": "^6.0.0",
    "typescript": "^5.7.0",
    "@types/react": "^19.0.0",
    "tailwindcss": "^4.0.0"              // Utility CSS (or use CSS modules)
  }
}
```

### 8.4 *arr-Style UI Components

```tsx
// components/Layout/Sidebar.tsx
export function Sidebar() {
  const links = [
    { icon: <Home />, label: 'Dashboard', path: '/' },
    { icon: <Bell />, label: 'Channels', path: '/channels' },
    { icon: <History />, label: 'History', path: '/history' },
    { icon: <Users />, label: 'Users', path: '/users' },
    { icon: <Settings />, label: 'Settings', path: '/settings' },
  ];

  return (
    <aside className="sidebar">
      <div className="sidebar-logo">
        <img src="/logo.svg" alt="Messagarr" />
        <span>Messagarr</span>
      </div>
      <nav className="sidebar-nav">
        {links.map(link => (
          <NavLink key={link.path} to={link.path} className="sidebar-link">
            {link.icon}
            <span>{link.label}</span>
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}

// pages/Channels.tsx
export function ChannelsPage() {
  const { data: channels, isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: () => api.channels.list()
  });

  const [showModal, setShowModal] = useState(false);

  return (
    <div className="page">
      <header className="page-header">
        <h1>Notification Channels</h1>
        <Button onClick={() => setShowModal(true)}>
          <Plus /> Add Channel
        </Button>
      </header>

      <div className="card-grid">
        {channels?.map(channel => (
          <ChannelCard key={channel.id} channel={channel} />
        ))}
      </div>

      {showModal && (
        <ChannelModal onClose={() => setShowModal(false)} />
      )}
    </div>
  );
}
```

### 8.5 Dark Theme CSS Variables

```css
/* styles/globals.css */
:root {
  /* *arr-inspired dark theme */
  --color-bg-primary: #1e2022;
  --color-bg-secondary: #282a2d;
  --color-bg-card: #323438;
  --color-bg-input: #3a3c40;

  --color-text-primary: #f5f5f5;
  --color-text-secondary: #9e9e9e;
  --color-text-muted: #6e6e6e;

  --color-accent: #3b82f6;
  --color-accent-hover: #2563eb;

  --color-success: #22c55e;
  --color-warning: #f59e0b;
  --color-error: #ef4444;

  --color-border: #404246;

  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;

  --sidebar-width: 220px;
}

body {
  background-color: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-family: 'Inter', -apple-system, sans-serif;
}

.card {
  background-color: var(--color-bg-card);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1rem;
}
```

---

## 9. Migration Path

### Phase 1: Core API (Week 1-2)

1. Set up solution structure with EF Core + SQLite
2. Implement entities and migrations
3. Port notification dispatch logic
4. Port resilience (Polly retry, deduplication)
5. Implement `/notify`, `/health`, `/ready`, `/status` endpoints
6. Add Prometheus metrics

**Validation**: All existing Go API tests pass against .NET version

### Phase 2: Auth & Persistence (Week 2-3)

1. Implement API key authentication
2. Implement cookie authentication for UI
3. Add first-launch middleware
4. Port channel configuration from YAML to DB seeding
5. Add notification logging to SQLite

### Phase 3: Management API (Week 3-4)

1. CRUD endpoints for channels
2. CRUD endpoints for priority groups
3. User management endpoints
4. Settings endpoints
5. Channel test endpoint

### Phase 4: Web UI (Week 4-6)

1. React app scaffolding with Vite
2. Initialize/setup wizard page
3. Login page
4. Dashboard with stats
5. Channel management (list, add, edit, delete, test)
6. Notification history
7. User management
8. Settings page

### Phase 5: Polish & Deploy (Week 6-7)

1. Docker image with multi-stage build
2. Migration tool (import from Go YAML config)
3. Documentation update
4. CI/CD pipeline

---

## 10. Docker & Deployment

### 10.1 Multi-Stage Dockerfile

```dockerfile
# Build React UI
FROM node:22-alpine AS ui-build
WORKDIR /app/ClientApp
COPY src/Messagarr/ClientApp/package*.json ./
RUN npm ci
COPY src/Messagarr/ClientApp/ ./
RUN npm run build

# Build .NET
FROM mcr.microsoft.com/dotnet/sdk:10.0-alpine AS dotnet-build
WORKDIR /src
COPY src/Messagarr/*.csproj ./Messagarr/
RUN dotnet restore ./Messagarr/Messagarr.csproj
COPY src/Messagarr/ ./Messagarr/
COPY --from=ui-build /app/ClientApp/dist ./Messagarr/wwwroot/
RUN dotnet publish ./Messagarr/Messagarr.csproj -c Release -o /app

# Runtime
FROM mcr.microsoft.com/dotnet/aspnet:10.0-alpine
WORKDIR /app
COPY --from=dotnet-build /app ./

# Create data directory for SQLite
RUN mkdir -p /config
ENV ConnectionStrings__Default="Data Source=/config/messagarr.db"

EXPOSE 4545
ENTRYPOINT ["dotnet", "Messagarr.dll"]
```

### 10.2 docker-compose.yml

```yaml
version: "3.8"
services:
  messagarr:
    image: messagarr:latest
    container_name: messagarr
    ports:
      - "4545:4545"
    volumes:
      - ./config:/config
    environment:
      - ASPNETCORE_ENVIRONMENT=Production
      - Logging__LogLevel__Default=Information
    restart: unless-stopped
```

### 10.3 appsettings.json

```json
{
  "ConnectionStrings": {
    "Default": "Data Source=/config/messagarr.db"
  },
  "Kestrel": {
    "Endpoints": {
      "Http": {
        "Url": "http://0.0.0.0:4545"
      }
    }
  },
  "Deduplication": {
    "TtlMinutes": 5
  },
  "Resilience": {
    "MaxRetries": 3,
    "BaseDelaySeconds": 1,
    "MaxDelaySeconds": 30
  },
  "Serilog": {
    "MinimumLevel": {
      "Default": "Information",
      "Override": {
        "Microsoft.AspNetCore": "Warning",
        "Microsoft.EntityFrameworkCore": "Warning"
      }
    },
    "WriteTo": [
      { "Name": "Console" }
    ]
  }
}
```

---

## 11. Development Workflow

### 11.1 Makefile

```makefile
.PHONY: build run test clean docker ui

# Default target
all: build

# Build everything
build: ui-build dotnet-build

# Build .NET
dotnet-build:
 dotnet build src/Messagarr/Messagarr.csproj

# Build UI
ui-build:
 cd src/Messagarr/ClientApp && npm ci && npm run build

# Run development (hot reload)
run:
 dotnet watch run --project src/Messagarr/Messagarr.csproj

# Run UI development server
ui-dev:
 cd src/Messagarr/ClientApp && npm run dev

# Run tests
test:
 dotnet test tests/Messagarr.Tests/Messagarr.Tests.csproj

# Run tests with coverage
coverage:
 dotnet test tests/Messagarr.Tests/Messagarr.Tests.csproj --collect:"XPlat Code Coverage"

# Clean build artifacts
clean:
 dotnet clean
 rm -rf src/Messagarr/wwwroot/*
 rm -rf src/Messagarr/ClientApp/dist

# EF Core migrations
migration-add:
 dotnet ef migrations add $(name) --project src/Messagarr/Messagarr.csproj

migration-apply:
 dotnet ef database update --project src/Messagarr/Messagarr.csproj

# Docker
docker-build:
 docker build -t messagarr:latest .

docker-run:
 docker run -d -p 4545:4545 -v $(PWD)/config:/config messagarr:latest

# Lint
lint:
 dotnet format src/Messagarr/Messagarr.csproj --verify-no-changes

# Format
format:
 dotnet format src/Messagarr/Messagarr.csproj

# Full development setup
dev-setup:
 dotnet restore
 cd src/Messagarr/ClientApp && npm ci
 dotnet ef database update --project src/Messagarr/Messagarr.csproj
```

### 11.2 VS Code Tasks

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "command": "make",
      "args": ["build"],
      "group": { "kind": "build", "isDefault": true }
    },
    {
      "label": "run",
      "command": "make",
      "args": ["run"],
      "group": "none",
      "isBackground": true
    },
    {
      "label": "test",
      "command": "make",
      "args": ["test"],
      "group": { "kind": "test", "isDefault": true }
    }
  ]
}
```

---

## 12. Implementation Phases

### Phase Summary

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| **Phase 1: Core API** | 2 weeks | `/notify`, `/health`, `/ready`, `/status`, metrics |
| **Phase 2: Auth & Data** | 1 week | API keys, cookies, SQLite, first-launch |
| **Phase 3: Management API** | 1 week | CRUD for channels, users, settings |
| **Phase 4: Web UI** | 2 weeks | Full React SPA with all pages |
| **Phase 5: Polish** | 1 week | Docker, docs, CI/CD, migration tool |

### Go → .NET Feature Mapping Checklist

| Go Feature | .NET Implementation | Status |
|------------|---------------------|--------|
| `POST /notify` | `NotificationEndpoints.SendNotification` | ⬜ |
| `GET /health` | `HealthEndpoints` (anonymous) | ⬜ |
| `GET /ready` | `HealthEndpoints` with DB check | ⬜ |
| `GET /status` | `HealthEndpoints` with stats | ⬜ |
| `GET /metrics` | prometheus-net middleware | ⬜ |
| `GET /swagger/*` | Swashbuckle | ⬜ |
| Deduplicator | `IMemoryCache` + SHA256 | ⬜ |
| Retrier | Polly `RetryStrategyOptions` | ⬜ |
| RateLimiter | Polly `SlidingWindowRateLimiter` | ⬜ |
| Discord channel | `DiscordChannel.cs` | ⬜ |
| Slack channel | `SlackChannel.cs` | ⬜ |
| Teams channel | `TeamsChannel.cs` | ⬜ |
| SMTP channel | `SmtpChannel.cs` with MailKit | ⬜ |
| YAML config | SQLite + UI | ⬜ |
| Env interpolation | N/A (use `appsettings.json` + env vars) | ⬜ |
| Priority groups | `PriorityGroup` entity + routing | ⬜ |
| Logging (slog) | Serilog | ⬜ |

---

## Appendix A: YAML Config Migration Tool

```csharp
// Services/Migration/YamlConfigImporter.cs
public class YamlConfigImporter
{
    public async Task ImportAsync(string yamlPath, MessagarrDbContext db)
    {
        var yaml = await File.ReadAllTextAsync(yamlPath);
        var deserializer = new DeserializerBuilder().Build();
        var config = deserializer.Deserialize<GoConfig>(yaml);

        // Import channels
        foreach (var channel in config.Channels)
        {
            db.Channels.Add(new Channel
            {
                Name = channel.Name ?? channel.Type,
                Type = Enum.Parse<ChannelType>(channel.Type, ignoreCase: true),
                WebhookUrl = channel.WebhookUrl,
                SmtpHost = channel.Host,
                SmtpPort = channel.Port,
                // ... etc
            });
        }

        // Import priority groups
        foreach (var (name, channelNames) in config.PriorityGroups)
        {
            var group = new PriorityGroup { Name = name };
            db.PriorityGroups.Add(group);
            // Link channels after SaveChanges
        }

        await db.SaveChangesAsync();
    }
}
```

---

## Appendix B: API Compatibility Layer

For drop-in replacement compatibility, ensure the `/notify` request/response format matches exactly:

**Request (unchanged)**:

```json
{
  "title": "Deployment Complete",
  "body": "Application v1.2.3 deployed",
  "priority": "high",
  "service": "deployment-service",
  "event_type": "deploy.success",
  "metadata": { "env": "prod" },
  "channels": ["discord", "email"]
}
```

**Response (unchanged)**:

```json
{
  "results": {
    "discord": { "success": true },
    "email": { "success": true }
  },
  "message": "Notification sent",
  "duration": "245ms",
  "success": true
}
```

---

## Next Steps

1. **Create the .NET solution** - Run `dotnet new sln -n Messagarr` and scaffold projects
2. **Set up EF Core** - Create DbContext and initial migration
3. **Port notification dispatch** - Start with Discord channel as reference
4. **Add basic auth** - API key handler + first-launch middleware
5. **Scaffold React app** - Vite + basic routing + login page

Would you like me to generate any specific component in detail?
