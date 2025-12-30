using Messagarr.Core.Entities;
using Microsoft.EntityFrameworkCore;

namespace Messagarr.Data;

public class MessagearrDbContext : DbContext
{
    public MessagearrDbContext(DbContextOptions<MessagearrDbContext> options) : base(options)
    {
    }

    public DbSet<User> Users => Set<User>();
    public DbSet<ApiKey> ApiKeys => Set<ApiKey>();
    public DbSet<Channel> Channels => Set<Channel>();
    public DbSet<NotificationHistory> NotificationHistories => Set<NotificationHistory>();
    public DbSet<NotificationChannelResult> NotificationChannelResults => Set<NotificationChannelResult>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        // User configuration
        modelBuilder.Entity<User>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.Username).IsUnique();
            entity.Property(e => e.Username).IsRequired().HasMaxLength(100);
            entity.Property(e => e.PasswordHash).IsRequired();
            entity.Property(e => e.CreatedAt).IsRequired();
            
            entity.HasMany(e => e.ApiKeys)
                .WithOne(e => e.User)
                .HasForeignKey(e => e.UserId)
                .OnDelete(DeleteBehavior.Cascade);
        });

        // ApiKey configuration
        modelBuilder.Entity<ApiKey>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.Key).IsUnique();
            entity.Property(e => e.Key).IsRequired();
            entity.Property(e => e.Name).IsRequired().HasMaxLength(200);
            entity.Property(e => e.CreatedAt).IsRequired();
        });

        // Channel configuration
        modelBuilder.Entity<Channel>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.Name).IsUnique();
            entity.Property(e => e.Name).IsRequired().HasMaxLength(100);
            entity.Property(e => e.Type).IsRequired();
            entity.Property(e => e.Configuration).IsRequired();
            entity.Property(e => e.CreatedAt).IsRequired();
        });

        // NotificationHistory configuration
        modelBuilder.Entity<NotificationHistory>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.Property(e => e.Title).IsRequired().HasMaxLength(500);
            entity.Property(e => e.Body).IsRequired();
            entity.Property(e => e.Priority).IsRequired().HasMaxLength(20);
            entity.Property(e => e.CreatedAt).IsRequired();
            
            entity.HasMany(e => e.ChannelResults)
                .WithOne(e => e.NotificationHistory)
                .HasForeignKey(e => e.NotificationHistoryId)
                .OnDelete(DeleteBehavior.Cascade);
        });

        // NotificationChannelResult configuration
        modelBuilder.Entity<NotificationChannelResult>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.Property(e => e.ChannelName).IsRequired().HasMaxLength(100);
        });
    }
}
