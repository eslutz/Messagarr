using Messagarr.Core.Entities;
using Messagarr.Core.Models;

namespace Messagarr.Core.Interfaces;

public interface IChannelService
{
    Task<IEnumerable<ChannelDto>> GetAllChannelsAsync();
    Task<ChannelDto?> GetChannelByIdAsync(int id);
    Task<ChannelDto> CreateChannelAsync(CreateChannelRequest request);
    Task<ChannelDto?> UpdateChannelAsync(int id, UpdateChannelRequest request);
    Task<bool> DeleteChannelAsync(int id);
    Task<(bool Success, string? Error)> TestChannelAsync(int id, TestChannelRequest request);
}
