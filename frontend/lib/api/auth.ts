import type { RegisterFormData } from '@/lib/validations/auth';

// Backend API base URL
const API_BASE = process.env.NEXT_PUBLIC_BACKEND_URL || 'http://localhost:8080';

/**
 * Registration response from backend
 */
export interface RegisterResponse {
    status: string;
    data: {
        userId: string;
        email: string;
        username: string;
        message: string;
    };
}

/**
 * Validation error from backend
 */
export interface ValidationError {
    field: string;
    message: string;
}

/**
 * Error response from backend
 */
export interface RegisterError {
    status: string;
    message: string;
    errors?: ValidationError[];
}

/**
 * Auth API service
 * Handles authentication-related API calls to the Go backend
 */
export const authApi = {
    /**
     * Register a new user
     * POST /api/auth/register
     */
    register: async (data: RegisterFormData): Promise<RegisterResponse> => {
        // Remove confirmPassword and agreeTerms before sending to backend
        const { _confirmPassword, _agreeTerms, ...registrationData } = data;

        const res = await fetch(`${API_BASE}/api/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(registrationData),
        });

        const responseData = await res.json();

        if (!res.ok) {
            const error: RegisterError = responseData;
            // Create error with additional properties for handling
            const err = new Error(error.message || 'Registration failed') as Error & {
                status: number;
                errors?: ValidationError[];
            };
            err.status = res.status;
            err.errors = error.errors;
            throw err;
        }

        return responseData as RegisterResponse;
    },

    /**
     * Check if email is available
     * Returns true if email is available, false if taken
     */
    checkEmailAvailability: async (email: string): Promise<boolean> => {
        try {
            const res = await fetch(`${API_BASE}/api/auth/check-email?email=${encodeURIComponent(email)}`);
            return res.ok;
        } catch {
            return true; // Assume available if check fails
        }
    },

    /**
     * Check if username is available
     * Returns true if username is available, false if taken
     */
    checkUsernameAvailability: async (username: string): Promise<boolean> => {
        try {
            const res = await fetch(
                `${API_BASE}/api/auth/check-username?username=${encodeURIComponent(username)}`
            );
            return res.ok;
        } catch {
            return true; // Assume available if check fails
        }
    },

    /**
     * Request password reset email
     * POST /api/auth/forgot-password
     */
    forgotPassword: async (email: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/api/auth/forgot-password`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email }),
        });

        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.message || 'Failed to request password reset');
        }
    },

    /**
     * Reset password using token
     * POST /api/auth/reset-password
     */
    resetPassword: async (token: string, newPassword: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/api/auth/reset-password`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ token, newPassword }),
        });

        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.message || 'Failed to reset password');
        }
    },

    /**
     * Validate reset token
     * GET /api/auth/validate-reset-token?token=<token>
     */
    validateResetToken: async (token: string): Promise<boolean> => {
        try {
            const res = await fetch(
                `${API_BASE}/api/auth/validate-reset-token?token=${encodeURIComponent(token)}`
            );
            return res.ok;
        } catch {
            return false;
        }
    },

    /**
     * Change password for authenticated user
     * POST /api/auth/change-password
     * Requires authentication
     */
    changePassword: async (currentPassword: string, newPassword: string): Promise<void> => {
        const res = await fetch(`${API_BASE}/api/auth/change-password`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include', // Include cookies for authentication
            body: JSON.stringify({ currentPassword, newPassword }),
        });

        if (!res.ok) {
            const error = await res.json();
            throw new Error(error.message || 'Failed to change password');
        }
    },
};
