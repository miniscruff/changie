//go:build race

package then

// delayTimeMS is the time between writes in DelayWrite.
// The race detector slows huh enough that moving to the next field can land
// after input written 50ms later, sending that input to the previous field.
const delayTimeMS = 200
