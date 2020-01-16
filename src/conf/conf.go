package conf

const BigEndian = false
const HEADER = "DNY"
const HeadLengthSize = 4

// 每个消息(包括头部)的最大长度， 这里最大可以设置4G
const BufferLength = 1024 * 70
const ChanMsgCount = 2
