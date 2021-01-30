# go-cqrs
CQRS library for golang

# Model
* App - bounded context with self Aggregates/Commands/Events/Views and infrastructure settings, can be joined with another App.

* Aggregate - commands handler, state owner, transaction boundary, events source
* Command - business request that changes aggregate
* Event - business event about aggregate changed

* View - query handler, events subscriber, eventual consistent with aggregates state 
* Query - business request to get some data from the view

* Saga - command handler to perform long async cross-aggregate business case.  
  Saga should be started automatic by some trigger event consumed. Or maybe by command???  
  Listen for events and dispatch commands.  
  Saga has non-business identity and state, usually implemented as finite state machine.  
  Maybe saga should handle some queries to describe self state? Most probably not, as state is non-business.  
  

# TODO
* ~~What to do if command not change state - not emit any events. Command validation error for example?~~  
  Return synchronous error from App.Command method and no events should be published. Maybe it is possible both error and events?
* Should event/command struct contains source/target aggregate id inside, or only payload wrapped with struct contains id?
* ~~How to create aggregates?~~ Allow to sent command to aggregate without ID, in this case id may be calculated during handling and will be in result events.
* How to implements cross aggregate cases? Saga? Process manager? DSL?
* Implement ES/DB dispatcher backends.  
  In case of DB, save new state of aggregate after command and notify all subscribers about events without persisting (sync/async?).
* What to do with events, needed in small scope without business value? Separate store with GC? Separate App instance?
* It is not recommended to call queries from aggregate handler, why?  
  Maybe it is OK if aggregate know about eventual consistency of view?
* Add aggregate version = count of applied events, for optimistic locking.  
  When publishing events after a handler executed, atomically check and update version (offset) of stream and retry command handler on error.  
  Version 0 = aggregate creation, no events applied in the past.  
  So commands should be free from side effects for safe retries.
* Define minimal API to event store interface.  
  Implement in-memory event store with concurrency support.
* Add BDD testing helper. Given/When/Then tests for the App. 