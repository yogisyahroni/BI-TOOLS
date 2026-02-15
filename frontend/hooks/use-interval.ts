import { useEffect, useRef } from 'react';

/**
 * Custom hook for setting up an interval with automatic cleanup
 * 
 * @param callback - Function to execute on each interval
 * @param delay - Delay in milliseconds (null to pause)
 */
export function useInterval(callback: () => void, delay: number | null) {
    const savedCallback = useRef(callback);

    // Remember the latest callback
    useEffect(() => {
        savedCallback.current = callback;
    }, [callback]);

    // Set up the interval
    useEffect(() => {
        // Don't schedule if no delay is specified
        if (delay === null) {
            return;
        }

        const id = setInterval(() => savedCallback.current(), delay);

        return () => clearInterval(id);
    }, [delay]);
}
