package btdemo

/*

#include <stdint.h>
#include <stdbool.h>

// The gateway function
bool onDataReceived_cgo(uint8_t *data,
                        uint8_t num_bytes,
                        uint32_t src_addr,
                        uint32_t dst_addr,
                        unsigned int qos,
                        uint8_t src_ep,
                        uint8_t dst_ep,
                        uint32_t travel_time,
                        uint8_t hop_count,
                        unsigned long long timestamp_ms_epoch)
{
	bool onDataReceived(uint8_t *data,
                        uint8_t num_bytes,
                        uint32_t src_addr,
                        uint32_t dst_addr,
                        unsigned int qos,
                        uint8_t src_ep,
                        uint8_t dst_ep,
                        uint32_t travel_time,
                        uint8_t hop_count,
                        unsigned long long timestamp_ms_epoch);
    return onDataReceived(data, num_bytes, src_addr, dst_addr, qos, src_ep, dst_ep, travel_time, hop_count, timestamp_ms_epoch);
}
*/
import "C"
