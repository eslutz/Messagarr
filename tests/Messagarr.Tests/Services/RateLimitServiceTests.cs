using Messagarr.Api.Services;

namespace Messagarr.Tests.Services;

public class RateLimitServiceTests
{
    [Fact]
    public async Task TryAcquireAsync_ShouldSucceed_ForFirstRequest()
    {
        // Arrange
        var service = new RateLimitService(tokensPerSecond: 10, bucketSize: 20);

        // Act
        var result = await service.TryAcquireAsync("channel1");

        // Assert
        Assert.True(result);
    }

    [Fact]
    public async Task TryAcquireAsync_ShouldFail_WhenBucketEmpty()
    {
        // Arrange
        var service = new RateLimitService(tokensPerSecond: 1, bucketSize: 2);

        // Act - Drain the bucket
        await service.TryAcquireAsync("channel1");
        await service.TryAcquireAsync("channel1");
        var result = await service.TryAcquireAsync("channel1");

        // Assert
        Assert.False(result);
    }

    [Fact]
    public async Task TryAcquireAsync_ShouldRefillOverTime()
    {
        // Arrange
        var service = new RateLimitService(tokensPerSecond: 10, bucketSize: 2);

        // Act - Drain the bucket
        await service.TryAcquireAsync("channel1");
        await service.TryAcquireAsync("channel1");
        
        // Wait for refill
        await Task.Delay(200);
        
        var result = await service.TryAcquireAsync("channel1");

        // Assert
        Assert.True(result);
    }

    [Fact]
    public async Task TryAcquireAsync_ShouldBeIndependent_ForDifferentChannels()
    {
        // Arrange
        var service = new RateLimitService(tokensPerSecond: 1, bucketSize: 1);

        // Act
        await service.TryAcquireAsync("channel1");
        var result = await service.TryAcquireAsync("channel2");

        // Assert
        Assert.True(result);
    }
}
