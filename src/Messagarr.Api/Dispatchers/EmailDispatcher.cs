using System.Net;
using System.Net.Mail;
using Messagarr.Core.Interfaces;
using Messagarr.Core.Models;

namespace Messagarr.Api.Dispatchers;

public class EmailDispatcher : IChannelDispatcher
{
    private readonly string _smtpHost;
    private readonly int _smtpPort;
    private readonly string _smtpUser;
    private readonly string _smtpPassword;
    private readonly string _fromEmail;
    private readonly string _toEmail;

    public EmailDispatcher(string smtpHost, int smtpPort, string smtpUser, string smtpPassword, 
        string fromEmail, string toEmail)
    {
        _smtpHost = smtpHost;
        _smtpPort = smtpPort;
        _smtpUser = smtpUser;
        _smtpPassword = smtpPassword;
        _fromEmail = fromEmail;
        _toEmail = toEmail;
        Name = "Email";
    }

    public string Name { get; }

    public async Task<(bool Success, string? Error)> SendAsync(NotificationRequest request)
    {
        try
        {
            using var client = new SmtpClient(_smtpHost, _smtpPort)
            {
                EnableSsl = true,
                Credentials = new NetworkCredential(_smtpUser, _smtpPassword)
            };

            var mailMessage = new MailMessage
            {
                From = new MailAddress(_fromEmail),
                Subject = $"[{request.Priority.ToUpper()}] {request.Title}",
                Body = FormatEmailBody(request),
                IsBodyHtml = true
            };

            mailMessage.To.Add(_toEmail);

            await client.SendMailAsync(mailMessage);
            return (true, null);
        }
        catch (Exception ex)
        {
            return (false, $"Failed to send email notification: {ex.Message}");
        }
    }

    private static string FormatEmailBody(NotificationRequest request)
    {
        return $@"
<html>
<body style='font-family: Arial, sans-serif;'>
    <h2>{request.Title}</h2>
    <p>{request.Body}</p>
    <hr>
    <table>
        <tr>
            <td><strong>Priority:</strong></td>
            <td>{request.Priority}</td>
        </tr>
        <tr>
            <td><strong>Service:</strong></td>
            <td>{request.Service ?? "N/A"}</td>
        </tr>
        <tr>
            <td><strong>Event Type:</strong></td>
            <td>{request.EventType ?? "N/A"}</td>
        </tr>
        <tr>
            <td><strong>Time:</strong></td>
            <td>{DateTime.UtcNow:yyyy-MM-dd HH:mm:ss} UTC</td>
        </tr>
    </table>
</body>
</html>";
    }
}
