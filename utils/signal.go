package utils

import (
	"sync"

	"golang.org/x/exp/maps"
)

type Signal0 struct {
	m       sync.Mutex
	nextId  int
	signals map[int]func()
}

func NewSignal0() Signal0 {
	return Signal0{}
}

func (s *Signal0) AddSlot(f func()) func() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.signals == nil {
		s.signals = make(map[int]func())
	}

	id := s.nextId
	s.signals[id] = f
	s.nextId++

	return func() {
		delete(s.signals, id)
	}
}

func (s *Signal0) Emit() {
	s.m.Lock()
	signals := maps.Clone(s.signals)
	s.m.Unlock()

	for _, s := range signals {
		s()
	}
}

func (s *Signal0) Union(o *Signal0) Signal0 {
	s.m.Lock()
	o.m.Lock()
	defer func() {
		s.m.Unlock()
		o.m.Unlock()
	}()

	signals := maps.Clone(s.signals)
	maps.Copy(signals, o.signals)

	return Signal0{
		nextId:  Max(s.nextId, o.nextId),
		signals: signals,
	}
}

func (s *Signal0) Clone() Signal0 {
	s.m.Lock()
	defer s.m.Unlock()

	return Signal0{
		nextId:  s.nextId,
		signals: maps.Clone(s.signals),
	}
}

func (s *Signal0) Clear() {
	s.m.Lock()
	defer s.m.Unlock()

	s.signals = make(map[int]func())
}

type Signal1[T any] struct {
	m       sync.Mutex
	nextId  int
	signals map[int]func(arg T)
}

func NewSignal1[T any]() Signal1[T] {
	return Signal1[T]{
		signals: make(map[int]func(arg T)),
	}
}

func (s *Signal1[T]) AddSlot(f func(arg T)) func() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.signals == nil {
		s.signals = make(map[int]func(arg T))
	}

	id := s.nextId
	s.signals[id] = f
	s.nextId++

	return func() {
		delete(s.signals, id)
	}
}

func (s *Signal1[T]) Emit(arg T) {
	s.m.Lock()
	signals := maps.Clone(s.signals)
	s.m.Unlock()

	for _, s := range signals {
		s(arg)
	}
}

func (s *Signal1[T]) Union(o *Signal1[T]) Signal1[T] {
	s.m.Lock()
	o.m.Lock()
	defer func() {
		s.m.Unlock()
		o.m.Unlock()
	}()

	signals := maps.Clone(s.signals)
	maps.Copy(signals, o.signals)

	return Signal1[T]{
		nextId:  Max(s.nextId, o.nextId),
		signals: signals,
	}
}

func (s *Signal1[T]) Clone() Signal1[T] {
	s.m.Lock()
	defer s.m.Unlock()

	return Signal1[T]{
		nextId:  s.nextId,
		signals: maps.Clone(s.signals),
	}
}

func (s *Signal1[T]) Clear() {
	s.m.Lock()
	defer s.m.Unlock()

	s.signals = make(map[int]func(arg T))
}

type SignalVar struct {
	m       sync.Mutex
	nextId  int
	signals map[int]func(args ...any)
}

func NewSignalVar() SignalVar {
	return SignalVar{
		signals: make(map[int]func(args ...any)),
	}
}

func (s *SignalVar) AddSlot(f func(args ...any)) func() {
	s.m.Lock()
	defer s.m.Unlock()

	if s.signals == nil {
		s.signals = make(map[int]func(args ...any))
	}

	id := s.nextId
	s.signals[id] = f
	s.nextId++

	return func() {
		delete(s.signals, id)
	}
}

func (s *SignalVar) Emit(args ...any) {
	s.m.Lock()
	signals := maps.Clone(s.signals)
	s.m.Unlock()

	for _, s := range signals {
		s(args...)
	}
}

func (s *SignalVar) Union(o *SignalVar) SignalVar {
	s.m.Lock()
	o.m.Lock()
	defer func() {
		s.m.Unlock()
		o.m.Unlock()
	}()

	signals := maps.Clone(s.signals)
	maps.Copy(signals, o.signals)

	return SignalVar{
		nextId:  Max(s.nextId, o.nextId),
		signals: signals,
	}
}

func (s *SignalVar) Clone() SignalVar {
	s.m.Lock()
	defer s.m.Unlock()

	return SignalVar{
		nextId:  s.nextId,
		signals: maps.Clone(s.signals),
	}
}

func (s *SignalVar) Clear() {
	s.m.Lock()
	defer s.m.Unlock()

	s.signals = make(map[int]func(args ...any))
}
