using System.Collections.Concurrent;
using System.Security.Cryptography;
using System.Text;

namespace Messagarr.Api.Services;

/// <summary>
/// Deduplication service to prevent duplicate notifications within a TTL window
/// </summary>
public class DeduplicationService
{
    private readonly ConcurrentDictionary<string, DateTime> _cache = new();
    private readonly TimeSpan _ttl;

    public DeduplicationService(TimeSpan ttl)
    {
        _ttl = ttl;
    }

    public bool IsDuplicate(string title, string body, string priority)
    {
        CleanupExpired();
        
        var hash = ComputeHash(title, body, priority);
        var now = DateTime.UtcNow;

        if (_cache.TryGetValue(hash, out var lastSeen))
        {
            if (now - lastSeen < _ttl)
            {
                return true;
            }
        }

        _cache[hash] = now;
        return false;
    }

    private void CleanupExpired()
    {
        var now = DateTime.UtcNow;
        var expiredKeys = _cache
            .Where(kvp => now - kvp.Value >= _ttl)
            .Select(kvp => kvp.Key)
            .ToList();

        foreach (var key in expiredKeys)
        {
            _cache.TryRemove(key, out _);
        }
    }

    private static string ComputeHash(string title, string body, string priority)
    {
        var input = $"{title}|{body}|{priority}";
        var bytes = SHA256.HashData(Encoding.UTF8.GetBytes(input));
        return Convert.ToBase64String(bytes);
    }
}
