using Messagarr.Data;
using Microsoft.EntityFrameworkCore;

var builder = WebApplication.CreateBuilder(args);

// Add services to the container
builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen(options =>
{
    options.SwaggerDoc("v1", new() { Title = "Messagarr API", Version = "v1" });
});

// Configure database
var dataDirectory = builder.Configuration.GetValue<string>("DataDirectory") ?? 
    Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), "Messagarr");
Directory.CreateDirectory(dataDirectory);

var dbPath = Path.Combine(dataDirectory, "messagarr.db");
builder.Services.AddDbContext<MessagearrDbContext>(options =>
    options.UseSqlite($"Data Source={dbPath}"));

// CORS for web UI
builder.Services.AddCors(options =>
{
    options.AddDefaultPolicy(policy =>
    {
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

var app = builder.Build();

// Run migrations automatically
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<MessagearrDbContext>();
    db.Database.Migrate();
}

// Configure the HTTP request pipeline
if (app.Environment.IsDevelopment())
{
    app.UseSwagger();
    app.UseSwaggerUI();
}

app.UseCors();

// Health check endpoint
app.MapGet("/health", () => Results.Ok(new { status = "healthy", timestamp = DateTime.UtcNow }))
    .WithName("Health")
    .WithOpenApi();

// Readiness check endpoint
app.MapGet("/ready", () => Results.Ok(new { status = "ready", timestamp = DateTime.UtcNow }))
    .WithName("Ready")
    .WithOpenApi();

app.Run();
