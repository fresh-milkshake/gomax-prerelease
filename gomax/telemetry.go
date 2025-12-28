package gomax

import (
	"context"
	"math/rand"
	"time"

	"github.com/fresh-milkshake/gomax/enums"
	"github.com/fresh-milkshake/gomax/internal/payloads"
	"github.com/fresh-milkshake/gomax/internal/utils"
)

var (
	navigationScreens = []int{150, 350, 450, 100, 300, 408, 409, 410, 412}
	navigationGraph   = map[int][]int{
		150: {350, 100, 300, 450},
		350: {150, 408, 409, 410, 412},
		450: {150, 100, 300},
		100: {150, 450, 300},
		300: {150, 450},
		408: {350, 409, 410, 412},
		409: {350, 408, 410, 412},
		410: {350, 408, 409, 412},
		412: {350, 408, 409, 410},
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// telemetryLoop имитирует пользовательскую активность: отправляет cold start и навигационные события.
func (c *MaxClient) telemetryLoop(ctx context.Context) {
	defer c.bgWG.Done()

	if err := c.sendColdStart(ctx); err != nil {
		c.logger.Debug("Cold start telemetry failed", "err", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !c.isWsConnected() {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := c.sendRandomNavigation(ctx); err != nil {
			c.logger.Debug("Navigation telemetry failed", "err", err)
		}

		time.Sleep(randomNavigationDelay())
	}
}

func (c *MaxClient) isWsConnected() bool {
	c.connMu.RLock()
	defer c.connMu.RUnlock()
	return c.ws != nil && c.isConnected
}

func (c *MaxClient) nextActionID() int {
	c.telemetryMu.Lock()
	defer c.telemetryMu.Unlock()
	c.actionID++
	return c.actionID
}

func (c *MaxClient) sendColdStart(ctx context.Context) error {
	userID := c.getUserID()
	screenTo := navigationScreens[0]
	actionID := c.nextActionID()

	event := payloads.NavigationEventPayload{
		Event:  "COLD_START",
		Time:   time.Now().UnixMilli(),
		Type:   "NAV",
		UserID: userID,
		Params: payloads.NavigationEventParams{
			ActionID:   actionID,
			ScreenTo:   screenTo,
			ScreenFrom: intPtr(1),
			SourceID:   1,
			SessionID:  c.sessionID,
		},
	}
	return c.sendNavigationEvents(ctx, []payloads.NavigationEventPayload{event})
}

func (c *MaxClient) sendRandomNavigation(ctx context.Context) error {
	c.telemetryMu.Lock()
	from := c.currentScreen
	to := pickNextScreen(from)
	c.currentScreen = to
	c.actionID++
	actionID := c.actionID
	c.telemetryMu.Unlock()

	userID := c.getUserID()

	event := payloads.NavigationEventPayload{
		Event:  "NAV",
		Time:   time.Now().UnixMilli(),
		Type:   "NAV",
		UserID: userID,
		Params: payloads.NavigationEventParams{
			ActionID:   actionID,
			ScreenTo:   to,
			ScreenFrom: &from,
			SourceID:   1,
			SessionID:  c.sessionID,
		},
	}
	return c.sendNavigationEvents(ctx, []payloads.NavigationEventPayload{event})
}

func (c *MaxClient) sendNavigationEvents(ctx context.Context, events []payloads.NavigationEventPayload) error {
	pl := payloads.NavigationPayload{Events: events}
	payloadMap, err := utils.ToMap(pl)
	if err != nil {
		return err
	}
	c.sendAsync(ctx, enums.OpcodeLog, payloadMap)
	return nil
}

func (c *MaxClient) getUserID() int64 {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	if c.Me != nil {
		return c.Me.ID
	}
	return 0
}

func pickNextScreen(from int) int {
	next := navigationGraph[from]
	if len(next) == 0 {
		return navigationScreens[0]
	}
	return next[rand.Intn(len(next))]
}

func randomNavigationDelay() time.Duration {
	r := rand.Float64()
	switch {
	case r < 0.05:
		return randomRange(1000, 3000)
	case r < 0.15:
		return randomRange(300, 1000)
	case r < 0.30:
		return randomRange(60, 300)
	case r < 0.50:
		return randomRange(5, 60)
	default:
		return randomRange(5, 20)
	}
}

func randomRange(lowMs, highMs int) time.Duration {
	if highMs <= lowMs {
		return time.Duration(lowMs) * time.Millisecond
	}
	return time.Duration(rand.Intn(highMs-lowMs)+lowMs) * time.Millisecond
}

func intPtr(v int) *int {
	return &v
}
