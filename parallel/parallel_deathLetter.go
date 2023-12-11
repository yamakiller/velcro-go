package parallel

type deathLetterHandler struct {
	_parallelSystem *ParallelSystem
}

var _ Handler = &deathLetterHandler{}

type DeadLetterEvent struct {
	PID     *PID        // The invalid process, to which the message was sent
	Message interface{} // The message that could not be delivered
	Sender  *PID        // the process that sent the Message
}

func (dp *deathLetterHandler) postSysMessage(pid *PID, message interface{}) {
	dp._parallelSystem._eventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: message,
	})
}

func (dp *deathLetterHandler) postUsrMessage(pid *PID, message interface{}) {
	/*metricsSystem, ok := dp.actorSystem.Extensions.Get(extensionId).(*Metrics)
	if ok && metricsSystem.enabled {
		ctx := context.Background()
		if instruments := metricsSystem.metrics.Get(metrics.InternalActorMetrics); instruments != nil {
			labels := []attribute.KeyValue{
				attribute.String("address", dp.actorSystem.Address()),
				attribute.String("messagetype", strings.Replace(fmt.Sprintf("%T", message), "*", "", 1)),
			}

			instruments.DeadLetterCount.Add(ctx, 1, metric.WithAttributes(labels...))
		}
	}
	_, msg, sender := UnwrapEnvelope(message)
	dp.actorSystem.EventStream.Publish(&DeadLetterEvent{
		PID:     pid,
		Message: msg,
		Sender:  sender,
	})*/
}

func (dp *deathLetterHandler) Stop(pid *PID) {
	dp.postSysMessage(pid, stopMessage)
}

func (dp *deathLetterHandler) overloadUsrMessage() int {
	return 0
}
