// Copyright 2020 Matteo Paier.
// Use of this source code is governed by a MIT license that can be found in the LICENSE file.

// Package pubsub provides extremely lightweight primitives for Publisher/Subscribe communication between goroutines.
// It allows creation of Pub/Sub groups and hierarchical topics.
package pubsub

import "strings"

// PubSub is a group of Publishers and Subscribers.
type PubSub struct {
	subs []subscriber
}

// subscriber describes a single subscriber: the topic is interested to and the channel it is using.
type subscriber struct {
	topic   string
	channel chan string
}

// isParent returns true iff topic1 is an ancestor of topic2.
func isParent(topic1 string, topic2 string) bool {
	t1 := strings.Split(topic1, ".")
	t2 := strings.Split(topic2, ".")
	if len(t1) > len(t2) {
		return false
	}
	for i, t := range t1 {
		if t != t2[i] {
			return false
		}
	}
	return true
}

// Publish allows to write on a specific topic (and its descendants).
func (ps *PubSub) Publish(topic string, message string) {
	for _, sub := range ps.subs {
		if isParent(topic, sub.topic) {
			sub.channel <- message
		}
	}
}

// Subscribe allows for subscription to a specific topic. It returns a channel that can be used to get the messages from the topic specified. A topic can be a subtopic of another.
// E.g.
//  ps.Subscribe("topic1.subtopic")
// receives also the messages from topic1.
func (ps *PubSub) Subscribe(topic string) chan string {
	var nsub subscriber
	var channel chan string = make(chan string)
	nsub.topic = topic
	nsub.channel = channel
	ps.subs = append(ps.subs, nsub)
	return channel
}

// Unsubscribe allows for the removal of a subscriber's channel. It requires a channel created with Subscribe.
func (ps *PubSub) Unsubscribe(channel chan string) {
	for i, sub := range ps.subs {
		if sub.channel == channel {
			close(sub.channel)
			ps.subs = append(ps.subs[:i], ps.subs[i+1:]...)
		}
	}
}
