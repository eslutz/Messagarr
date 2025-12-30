namespace Messagarr.Core.Models;

public class LoginRequest
{
    public string Username { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
}

public class LoginResponse
{
    public bool Success { get; set; }
    public string? Token { get; set; }
    public string? ApiKey { get; set; }
    public string? Message { get; set; }
}

public class SetupRequest
{
    public string Username { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
}

public class SetupResponse
{
    public bool Success { get; set; }
    public string? ApiKey { get; set; }
    public string? Message { get; set; }
}
