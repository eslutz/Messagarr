using System.Collections.Concurrent;

namespace Messagarr.Api.Services;

/// <summary>
/// Rate limiting service using token bucket algorithm
/// </summary>
public class RateLimitService
{
    private readonly ConcurrentDictionary<string, TokenBucket> _buckets = new();
    private readonly int _tokensPerSecond;
    private readonly int _bucketSize;

    public RateLimitService(int tokensPerSecond = 10, int bucketSize = 20)
    {
        _tokensPerSecond = tokensPerSecond;
        _bucketSize = bucketSize;
    }

    public async Task<bool> TryAcquireAsync(string channelId)
    {
        var bucket = _buckets.GetOrAdd(channelId, _ => new TokenBucket(_tokensPerSecond, _bucketSize));
        return await bucket.TryAcquireAsync();
    }

    private class TokenBucket
    {
        private readonly int _tokensPerSecond;
        private readonly int _bucketSize;
        private double _tokens;
        private DateTime _lastRefill;
        private readonly SemaphoreSlim _lock = new(1, 1);

        public TokenBucket(int tokensPerSecond, int bucketSize)
        {
            _tokensPerSecond = tokensPerSecond;
            _bucketSize = bucketSize;
            _tokens = bucketSize;
            _lastRefill = DateTime.UtcNow;
        }

        public async Task<bool> TryAcquireAsync()
        {
            await _lock.WaitAsync();
            try
            {
                Refill();

                if (_tokens >= 1)
                {
                    _tokens -= 1;
                    return true;
                }

                return false;
            }
            finally
            {
                _lock.Release();
            }
        }

        private void Refill()
        {
            var now = DateTime.UtcNow;
            var elapsed = (now - _lastRefill).TotalSeconds;
            var tokensToAdd = elapsed * _tokensPerSecond;

            _tokens = Math.Min(_bucketSize, _tokens + tokensToAdd);
            _lastRefill = now;
        }
    }
}
