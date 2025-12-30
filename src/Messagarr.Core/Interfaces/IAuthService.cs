using Messagarr.Core.Models;

namespace Messagarr.Core.Interfaces;

public interface IAuthService
{
    Task<SetupResponse> SetupInitialUserAsync(SetupRequest request);
    Task<LoginResponse> LoginAsync(LoginRequest request);
    Task<bool> IsSetupCompleteAsync();
    Task<string> GenerateApiKeyAsync(int userId, string name);
    Task<bool> ValidateApiKeyAsync(string apiKey);
}
