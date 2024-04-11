namespace go protocols.prvs

struct ClientRequestMessage{
    1:binary data;
    2:i64 timeout;
}