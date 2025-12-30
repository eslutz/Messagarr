using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Security.Cryptography;
using System.Text;
using Messagarr.Core.Entities;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;
using Messagarr.Data;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.IdentityModel.Tokens;

namespace Messagarr.Api.Services;

public class AuthService : IAuthService
{
    private readonly MessagearrDbContext _dbContext;
    private readonly IConfiguration _configuration;

    public AuthService(MessagearrDbContext dbContext, IConfiguration configuration)
    {
        _dbContext = dbContext;
        _configuration = configuration;
    }

    public async Task<bool> IsSetupCompleteAsync()
    {
        return await _dbContext.Users.AnyAsync();
    }

    public async Task<SetupResponse> SetupInitialUserAsync(SetupRequest request)
    {
        if (await IsSetupCompleteAsync())
        {
            return new SetupResponse
            {
                Success = false,
                Message = "Setup has already been completed"
            };
        }

        if (string.IsNullOrWhiteSpace(request.Username) || string.IsNullOrWhiteSpace(request.Password))
        {
            return new SetupResponse
            {
                Success = false,
                Message = "Username and password are required"
            };
        }

        if (request.Password.Length < 8)
        {
            return new SetupResponse
            {
                Success = false,
                Message = "Password must be at least 8 characters long"
            };
        }

        var user = new User
        {
            Username = request.Username,
            PasswordHash = HashPassword(request.Password),
            CreatedAt = DateTime.UtcNow,
            IsActive = true
        };

        _dbContext.Users.Add(user);
        await _dbContext.SaveChangesAsync();

        var apiKey = await GenerateApiKeyAsync(user.Id, "Default API Key");

        return new SetupResponse
        {
            Success = true,
            ApiKey = apiKey,
            Message = "Setup completed successfully"
        };
    }

    public async Task<LoginResponse> LoginAsync(LoginRequest request)
    {
        var user = await _dbContext.Users
            .FirstOrDefaultAsync(u => u.Username == request.Username && u.IsActive);

        if (user == null || !VerifyPassword(request.Password, user.PasswordHash))
        {
            return new LoginResponse
            {
                Success = false,
                Message = "Invalid username or password"
            };
        }

        user.LastLoginAt = DateTime.UtcNow;
        await _dbContext.SaveChangesAsync();

        var token = GenerateJwtToken(user);
        var apiKey = await _dbContext.ApiKeys
            .Where(k => k.UserId == user.Id && k.IsActive)
            .OrderByDescending(k => k.CreatedAt)
            .Select(k => k.Key)
            .FirstOrDefaultAsync();

        return new LoginResponse
        {
            Success = true,
            Token = token,
            ApiKey = apiKey,
            Message = "Login successful"
        };
    }

    public async Task<string> GenerateApiKeyAsync(int userId, string name)
    {
        var key = GenerateSecureApiKey();

        var apiKey = new ApiKey
        {
            Key = key,
            Name = name,
            UserId = userId,
            CreatedAt = DateTime.UtcNow,
            IsActive = true
        };

        _dbContext.ApiKeys.Add(apiKey);
        await _dbContext.SaveChangesAsync();

        return key;
    }

    public async Task<bool> ValidateApiKeyAsync(string apiKey)
    {
        var key = await _dbContext.ApiKeys
            .Include(k => k.User)
            .FirstOrDefaultAsync(k => k.Key == apiKey && k.IsActive);

        if (key == null || !key.User.IsActive)
        {
            return false;
        }

        if (key.ExpiresAt.HasValue && key.ExpiresAt.Value < DateTime.UtcNow)
        {
            return false;
        }

        key.LastUsedAt = DateTime.UtcNow;
        await _dbContext.SaveChangesAsync();

        return true;
    }

    private string HashPassword(string password)
    {
        using var sha256 = SHA256.Create();
        var hashedBytes = sha256.ComputeHash(Encoding.UTF8.GetBytes(password));
        return Convert.ToBase64String(hashedBytes);
    }

    private bool VerifyPassword(string password, string hash)
    {
        return HashPassword(password) == hash;
    }

    private string GenerateJwtToken(User user)
    {
        var jwtKey = _configuration["Jwt:Key"] ?? GenerateSecureApiKey();
        var securityKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtKey));
        var credentials = new SigningCredentials(securityKey, SecurityAlgorithms.HmacSha256);

        var claims = new[]
        {
            new Claim(JwtRegisteredClaimNames.Sub, user.Id.ToString()),
            new Claim(JwtRegisteredClaimNames.UniqueName, user.Username),
            new Claim(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
        };

        var token = new JwtSecurityToken(
            issuer: _configuration["Jwt:Issuer"] ?? "Messagarr",
            audience: _configuration["Jwt:Audience"] ?? "Messagarr",
            claims: claims,
            expires: DateTime.UtcNow.AddDays(7),
            signingCredentials: credentials
        );

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    private static string GenerateSecureApiKey()
    {
        var randomNumber = new byte[32];
        using var rng = RandomNumberGenerator.Create();
        rng.GetBytes(randomNumber);
        return Convert.ToBase64String(randomNumber).Replace("+", "").Replace("/", "").Replace("=", "");
    }
}
