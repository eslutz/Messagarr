namespace Messagarr.Api.Services;

/// <summary>
/// Retry service with exponential backoff
/// </summary>
public class RetryService
{
    private readonly int _maxRetries;
    private readonly TimeSpan _initialDelay;
    private readonly TimeSpan _maxDelay;

    public RetryService(int maxRetries = 3, int initialDelaySeconds = 1, int maxDelaySeconds = 30)
    {
        _maxRetries = maxRetries;
        _initialDelay = TimeSpan.FromSeconds(initialDelaySeconds);
        _maxDelay = TimeSpan.FromSeconds(maxDelaySeconds);
    }

    public async Task<T> ExecuteWithRetryAsync<T>(Func<Task<T>> operation)
    {
        Exception? lastException = null;

        for (int attempt = 0; attempt <= _maxRetries; attempt++)
        {
            try
            {
                return await operation();
            }
            catch (Exception ex)
            {
                lastException = ex;

                if (attempt == _maxRetries)
                {
                    break;
                }

                var delay = CalculateDelay(attempt);
                await Task.Delay(delay);
            }
        }

        throw new InvalidOperationException(
            $"Operation failed after {_maxRetries + 1} attempts", 
            lastException);
    }

    private TimeSpan CalculateDelay(int attempt)
    {
        var delay = TimeSpan.FromMilliseconds(_initialDelay.TotalMilliseconds * Math.Pow(2, attempt));
        return delay > _maxDelay ? _maxDelay : delay;
    }
}
