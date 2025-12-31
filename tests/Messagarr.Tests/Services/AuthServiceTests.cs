using Messagarr.Api.Services;
using Messagarr.Core.Entities;
using Messagarr.Data;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;

namespace Messagarr.Tests.Services;

public class AuthServiceTests : IDisposable
{
    private readonly MessagearrDbContext _context;
    private readonly AuthService _authService;

    public AuthServiceTests()
    {
        var options = new DbContextOptionsBuilder<MessagearrDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;

        _context = new MessagearrDbContext(options);

        var configuration = new ConfigurationBuilder()
            .AddInMemoryCollection(new Dictionary<string, string?>
            {
                {"Jwt:Key", "ThisIsATestKeyThatIsLongEnoughForHS256"},
                {"Jwt:Issuer", "TestIssuer"},
                {"Jwt:Audience", "TestAudience"}
            })
            .Build();

        _authService = new AuthService(_context, configuration);
    }

    [Fact]
    public async Task SetupInitialUser_ShouldCreateUser_WhenNoUserExists()
    {
        // Arrange
        var request = new Core.Models.SetupRequest
        {
            Username = "admin",
            Password = "password123"
        };

        // Act
        var result = await _authService.SetupInitialUserAsync(request);

        // Assert
        Assert.True(result.Success);
        Assert.NotNull(result.ApiKey);
        Assert.Equal("Setup completed successfully", result.Message);

        var user = await _context.Users.FirstOrDefaultAsync();
        Assert.NotNull(user);
        Assert.Equal("admin", user.Username);
    }

    [Fact]
    public async Task SetupInitialUser_ShouldFail_WhenUserAlreadyExists()
    {
        // Arrange
        _context.Users.Add(new User
        {
            Username = "existing",
            PasswordHash = "hash",
            CreatedAt = DateTime.UtcNow
        });
        await _context.SaveChangesAsync();

        var request = new Core.Models.SetupRequest
        {
            Username = "admin",
            Password = "password123"
        };

        // Act
        var result = await _authService.SetupInitialUserAsync(request);

        // Assert
        Assert.False(result.Success);
        Assert.Equal("Setup has already been completed", result.Message);
    }

    [Fact]
    public async Task Login_ShouldSucceed_WithValidCredentials()
    {
        // Arrange
        var setupRequest = new Core.Models.SetupRequest
        {
            Username = "testuser",
            Password = "password123"
        };
        await _authService.SetupInitialUserAsync(setupRequest);

        var loginRequest = new Core.Models.LoginRequest
        {
            Username = "testuser",
            Password = "password123"
        };

        // Act
        var result = await _authService.LoginAsync(loginRequest);

        // Assert
        Assert.True(result.Success);
        Assert.NotNull(result.Token);
        Assert.NotNull(result.ApiKey);
    }

    [Fact]
    public async Task Login_ShouldFail_WithInvalidCredentials()
    {
        // Arrange
        var setupRequest = new Core.Models.SetupRequest
        {
            Username = "testuser",
            Password = "password123"
        };
        await _authService.SetupInitialUserAsync(setupRequest);

        var loginRequest = new Core.Models.LoginRequest
        {
            Username = "testuser",
            Password = "wrongpassword"
        };

        // Act
        var result = await _authService.LoginAsync(loginRequest);

        // Assert
        Assert.False(result.Success);
        Assert.Equal("Invalid username or password", result.Message);
    }

    [Fact]
    public async Task IsSetupComplete_ShouldReturnFalse_WhenNoUserExists()
    {
        // Act
        var result = await _authService.IsSetupCompleteAsync();

        // Assert
        Assert.False(result);
    }

    [Fact]
    public async Task IsSetupComplete_ShouldReturnTrue_WhenUserExists()
    {
        // Arrange
        _context.Users.Add(new User
        {
            Username = "testuser",
            PasswordHash = "hash",
            CreatedAt = DateTime.UtcNow
        });
        await _context.SaveChangesAsync();

        // Act
        var result = await _authService.IsSetupCompleteAsync();

        // Assert
        Assert.True(result);
    }

    public void Dispose()
    {
        _context.Database.EnsureDeleted();
        _context.Dispose();
    }
}
