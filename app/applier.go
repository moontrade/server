package app

import "github.com/golang/snappy"

// runWriteApplier is a background routine that handles all write requests.
// Its job is to apply the request to the Raft log and returns the result to
// writeRequest.
func runWriteApplier(conf Config, m *machine, ra *raftWrap) {
	var maxReqs = 1024 // TODO: make configurable
	for {
		// Gather up as many requests (up to 256) into a single list.
		var reqs []*writeRequestFuture
		r := <-m.wrC
		reqs = append(reqs, r)
		var done bool
		for !done {
			select {
			case r := <-m.wrC:
				reqs = append(reqs, r)
				done = len(reqs) == maxReqs
			default:
				done = true
			}
		}
		// Combined multiple requests the data to a single, snappy-encoded,
		// message using the following binary format:
		// (count, cmd...)
		//   - count: uvarint
		//   - cmd: (count, args...)
		//     - count: uvarint
		//     - arg: (count, byte...)
		//       - count: uvarint
		var data []byte
		data = appendUvarint(data, uint64(len(reqs)))
		for _, r := range reqs {
			data = appendUvarint(data, uint64(len(r.args)))
			for _, arg := range r.args {
				data = appendUvarint(data, uint64(len(arg)))
				data = append(data, arg...)
			}
		}

		data = snappy.Encode(nil, data)

		// Apply the data and read back the messages
		resps, err := func() ([]applyResp, error) {
			// THE ONLY APPLY CALL IN THE CODEBASE SO ENJOY IT
			f := ra.Apply(data, 0)
			err := f.Error()
			if err != nil {
				return nil, err
			}
			return f.Response().([]applyResp), nil
		}()
		if err != nil {
			for _, r := range reqs {
				r.err = errRaftConvert(ra, err)
				r.wg.Done()
			}
		} else {
			for i := range reqs {
				reqs[i].resp = resps[i].resp
				reqs[i].elap = resps[i].elap
				reqs[i].err = resps[i].err
				reqs[i].wg.Done()
			}
		}
	}
}
