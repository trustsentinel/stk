package main

import (
	"fmt"
	"sync"
	"time"
)

// SafeTokens structure which persists the current tokens
type SafeTokens struct {
	tokens map[string]Registration
	mux    sync.Mutex
}

// Registration stores information of registration
type Registration struct {
	token   string
	updated time.Time
}

// Update add new token
func (c *SafeTokens) Update(token string) {
	c.mux.Lock()
	if _, ok := c.tokens[token]; ok {
		fmt.Println("token already exists")
	}
	c.tokens[token] = Registration{
		token:   token,
		updated: time.Now()}
	c.mux.Unlock()
}

// Remove token from general structure
func (c *SafeTokens) Remove(token string) {
	c.mux.Lock()
	if _, ok := c.tokens[token]; ok {
		delete(c.tokens, token)
	}
	c.mux.Unlock()
}

// ToS provides a string representation of Registration
func (r *Registration) ToS() string {
	return fmt.Sprintf(`{ "token": "%s", "updated": "%s" }`, r.token, r.updated)
}

// Get provides available tokens
func (c *SafeTokens) Get() []string {
	v := make([]string, 0, len(c.tokens))
	for _, r := range c.tokens {
		v = append(v, r.ToS())
	}
	return v
}
