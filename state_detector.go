package event_fsm

// StateDetector is a state detector
type StateDetector[T comparable] struct {
	states        map[string]*State[T]
	mainStateName StateName
}

func NewStateDetector[T comparable]() *StateDetector[T] {
	return &StateDetector[T]{
		states: make(map[string]*State[T]),
	}
}

func (sd *StateDetector[T]) NewState(name StateName, executor Executor[T], stateType StateType) *State[T] {
	if _, ok := _stateNames[name.String()]; !ok {
		panic(ErrStateNameNotFound)
	}

	state := &State[T]{
		Name:      name,
		Executor:  executor,
		Next:      make(map[string]*State[T]),
		StateType: stateType,
	}
	sd.states[name.String()] = state

	return state
}

func (sd *StateDetector[T]) SetMainState(state StateName) {
	sd.mainStateName = state
}

func (sd *StateDetector[T]) getMainState() (StateName, error) {
	if sd.mainStateName.String() == "" {
		return "", ErrMainStateNotFound
	}
	return sd.mainStateName, nil
}

func (sd *StateDetector[T]) stateByName(name StateName) (*State[T], error) {
	if state, ok := sd.states[name.String()]; ok {
		return state, nil
	}

	return nil, ErrStateNotFound
}

func (sd *StateDetector[T]) getNextState(state *State[T], response ResultStatus) (*State[T], bool) {
	if nextState, ok := state.Next[response.String()]; ok {

		if _, ok = sd.states[nextState.Name.String()]; ok {
			return nextState, true
		}
	}

	return nil, false
}
