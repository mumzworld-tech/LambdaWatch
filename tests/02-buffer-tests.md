# Buffer Tests

## 2.1 Basic Operations

### TC-2.1.1: Add Single Entry
- **Action**: `buffer.Add(entry)`
- **Expected**: `buffer.Len() == 1`

### TC-2.1.2: Add Batch
- **Action**: `buffer.AddBatch([]LogEntry{e1, e2, e3})`
- **Expected**: `buffer.Len() == 3`

### TC-2.1.3: Flush Partial
- **Setup**: Add 10 entries
- **Action**: `buffer.Flush(5)`
- **Expected**: Returns 5 entries, `buffer.Len() == 5`

### TC-2.1.4: Flush All
- **Setup**: Add 10 entries
- **Action**: `buffer.Flush(100)`
- **Expected**: Returns 10 entries, `buffer.Len() == 0`

### TC-2.1.5: Flush Empty Buffer
- **Action**: `buffer.Flush(10)` on empty buffer
- **Expected**: Returns nil

---

## 2.2 Bounded Size

### TC-2.2.1: At Capacity
- **Setup**: Buffer with maxSize=100, add 100 entries
- **Action**: Check `buffer.Len()`
- **Expected**: 100 entries

### TC-2.2.2: Overflow Drops Oldest
- **Setup**: Buffer with maxSize=100, add entries 1-100
- **Action**: Add entry 101
- **Expected**:
  - `buffer.Len() == 100`
  - Entry 1 dropped, entries 2-101 remain

### TC-2.2.3: Batch Overflow
- **Setup**: Buffer with maxSize=100, add 50 entries
- **Action**: `AddBatch()` with 75 entries
- **Expected**:
  - `buffer.Len() == 100`
  - Oldest 25 entries dropped

---

## 2.3 Byte Size Tracking

### TC-2.3.1: ByteSize Increases on Add
- **Setup**: Empty buffer
- **Action**: Add entry with 100-byte message
- **Expected**: `buffer.ByteSize() >= 100`

### TC-2.3.2: ByteSize Decreases on Flush
- **Setup**: Add entries totaling 1000 bytes
- **Action**: Flush half
- **Expected**: `buffer.ByteSize() ~= 500`

### TC-2.3.3: FlushBySize Respects Byte Limit
- **Setup**: Add 10 entries of 100 bytes each
- **Action**: `buffer.FlushBySize(100, 350)`
- **Expected**: Returns 3 entries (~300 bytes)

### TC-2.3.4: FlushBySize Single Large Entry
- **Setup**: Add 1 entry of 1000 bytes
- **Action**: `buffer.FlushBySize(100, 500)`
- **Expected**: Returns 1 entry (even though > maxBytes, at least 1 returned)

---

## 2.4 Thread Safety

### TC-2.4.1: Concurrent Add
- **Action**: 10 goroutines each adding 100 entries
- **Expected**: No race conditions, final count correct

### TC-2.4.2: Concurrent Add and Flush
- **Action**:
  - Goroutine 1: Continuously adding entries
  - Goroutine 2: Continuously flushing
- **Expected**: No race conditions, no panics

### TC-2.4.3: Concurrent Len and ByteSize
- **Action**: Call `Len()` and `ByteSize()` concurrently with Add/Flush
- **Expected**: No race conditions

---

## 2.5 Drain

### TC-2.5.1: Drain Returns All
- **Setup**: Add 50 entries
- **Action**: `buffer.Drain()`
- **Expected**: Returns 50 entries, buffer empty

### TC-2.5.2: Drain Closes Buffer
- **Action**: `buffer.Drain()`, then `buffer.Add(entry)`
- **Expected**: Add returns false, entry not added

### TC-2.5.3: Drain on Empty Buffer
- **Action**: `buffer.Drain()` on empty buffer
- **Expected**: Returns empty slice, buffer closed

---

## 2.6 Ready Signal

### TC-2.6.1: AddBatch Signals Ready
- **Setup**: Listener on `buffer.Ready()` channel
- **Action**: `buffer.AddBatch(entries)`
- **Expected**: Ready channel receives signal

### TC-2.6.2: SignalReady Manual
- **Setup**: Listener on `buffer.Ready()` channel
- **Action**: `buffer.SignalReady()`
- **Expected**: Ready channel receives signal

### TC-2.6.3: Ready Non-Blocking
- **Action**: Call `SignalReady()` twice immediately
- **Expected**: No blocking (channel has buffer of 1)

---

## 2.7 Entry Size Calculation

### TC-2.7.1: Size Includes All Fields
- **Setup**: Entry with Message="hello", Type="function", RequestID="abc-123"
- **Action**: `entry.Size()`
- **Expected**: `len("hello") + len("function") + len("abc-123") + 8 = 24`

### TC-2.7.2: Empty Fields
- **Setup**: Entry with only Message="test"
- **Action**: `entry.Size()`
- **Expected**: `len("test") + 8 = 12`
