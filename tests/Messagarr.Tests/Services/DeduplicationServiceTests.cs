using Messagarr.Api.Services;

namespace Messagarr.Tests.Services;

public class DeduplicationServiceTests
{
    [Fact]
    public void IsDuplicate_ShouldReturnFalse_ForFirstNotification()
    {
        // Arrange
        var service = new DeduplicationService(TimeSpan.FromMinutes(5));

        // Act
        var result = service.IsDuplicate("Test Title", "Test Body", "normal");

        // Assert
        Assert.False(result);
    }

    [Fact]
    public void IsDuplicate_ShouldReturnTrue_ForDuplicateWithinTTL()
    {
        // Arrange
        var service = new DeduplicationService(TimeSpan.FromMinutes(5));

        // Act
        var first = service.IsDuplicate("Test Title", "Test Body", "normal");
        var second = service.IsDuplicate("Test Title", "Test Body", "normal");

        // Assert
        Assert.False(first);
        Assert.True(second);
    }

    [Fact]
    public void IsDuplicate_ShouldReturnFalse_ForDifferentContent()
    {
        // Arrange
        var service = new DeduplicationService(TimeSpan.FromMinutes(5));

        // Act
        var first = service.IsDuplicate("Test Title 1", "Test Body 1", "normal");
        var second = service.IsDuplicate("Test Title 2", "Test Body 2", "normal");

        // Assert
        Assert.False(first);
        Assert.False(second);
    }

    [Fact]
    public void IsDuplicate_ShouldReturnFalse_ForSameContentDifferentPriority()
    {
        // Arrange
        var service = new DeduplicationService(TimeSpan.FromMinutes(5));

        // Act
        var first = service.IsDuplicate("Test Title", "Test Body", "normal");
        var second = service.IsDuplicate("Test Title", "Test Body", "high");

        // Assert
        Assert.False(first);
        Assert.False(second);
    }

    [Fact]
    public async Task IsDuplicate_ShouldReturnFalse_AfterTTLExpires()
    {
        // Arrange
        var service = new DeduplicationService(TimeSpan.FromMilliseconds(100));

        // Act
        var first = service.IsDuplicate("Test Title", "Test Body", "normal");
        await Task.Delay(150); // Wait for TTL to expire
        var second = service.IsDuplicate("Test Title", "Test Body", "normal");

        // Assert
        Assert.False(first);
        Assert.False(second);
    }
}
