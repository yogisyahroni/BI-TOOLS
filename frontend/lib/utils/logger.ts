export enum LogLevel {
    INFO = 'INFO',
    WARN = 'WARN',
    ERROR = 'ERROR',
    AUDIT = 'AUDIT'
}

export class ProductionLogger {
    /**
     * Standardized JSON logging for production monitoring (ELK/CloudWatch ready)
     */
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    static log(level: LogLevel, message: string, meta: any = {}) {
        const logEntry = {
            timestamp: new Date().toISOString(),
            level,
            message,
            ...meta,
            env: process.env.NODE_ENV || 'development'
        };

        if (process.env.NODE_ENV === 'production') {
            console.warn(JSON.stringify(logEntry));
        } else {
            // Readable format for dev
            console.warn(`[${level}] ${message}`, meta);
        }
    }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    static info(message: string, meta?: any) { this.log(LogLevel.INFO, message, meta); }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    static warn(message: string, meta?: any) { this.log(LogLevel.WARN, message, meta); }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    static error(message: string, meta?: any) { this.log(LogLevel.ERROR, message, meta); }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    static audit(message: string, meta?: any) { this.log(LogLevel.AUDIT, message, meta); }
}
