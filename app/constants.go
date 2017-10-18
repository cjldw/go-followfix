package app

// APP_CONFIG_PATH
const APP_CONFIG_PATH  = "conf/app.toml"

// DB_USERS_DATA
const DB_USERS_DATA  = "users_data"

// REDIS_SOCIAL social redis key name
const REDIS_SOCIAL  =  "social"

// USER_FOLLOW_TABLE_PREFIX
const USER_FOLLOW_TABLE_PREFIX  = "user_follow_"

// USER_FOLLOW_SPLIT_TABLE_NUM
const USER_FOLLOW_SPLIT_TABLE_NUM  =  10

// FRIEND_SYSTEM_USER_FOLLOW User Follow Sorted Set Collection
// Value is UID and Score is UID's Fans Number
const FRIEND_SYSTEM_USER_FOLLOW = "friend:system:user:follow:"

// FRIEND_SYSTEM_USER_FANS User's Fans Sorted Set Collection
// Values is UID and Score is UID's Fans Number
const FRIEND_SYSTEM_USER_FANS = "friend:system:user:fans:"

// FRIEND_SYSTEM_USER_FRIENDS User's Friends Sorted Set Collection
// Value is UID and Score is UID's Fans Number
const FRIEND_SYSTEM_USER_FRIENDS = "friend:system:user:friends:"

// TMP_UID_FANS_NUM storage uid fans number temporary
const TMP_UID_FANS_NUM  = "tmp:uid:fans:num"

// PROCESS_UID_VAVEL goroutine 阀门
const PROCESS_UID_VAVEL  =  10
