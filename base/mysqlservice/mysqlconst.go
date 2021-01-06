package mysqlservice

// READTIMEOUT 数据库读取超时时间
const READTIMEOUT = "30s"

// WRITETIMEOUT 数据库写入超时时间
const WRITETIMEOUT = "2m30s"

// MAXALLOWEDPACKET 最大允许包体大小
const MAXALLOWEDPACKET = 16 * 1024 * 1024 // 16 MiB
// const maxAllowedPacket = 4 * 1024 * 1024 // 4 MiB

// MAXPACKETNUM 最大的包体个数
const MAXPACKETNUM = 8

// const maxPacketNum = 8                    // 最大缓存包数量。不需要太多，占内存，且到不了这么多个包
