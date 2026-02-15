import { z } from 'zod';

/**
 * Email validation schema with business rules
 * - Must be valid email format
 * - Required field
 * - Max 255 characters (database constraint)
 * - Normalized to lowercase and trimmed
 */
export const emailSchema = z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address')
    .max(255, 'Email is too long')
    .transform((val) => val.toLowerCase().trim());

/**
 * Password validation schema (client-side feedback only)
 * Note: Real validation happens on the server
 * - Minimum 8 characters (industry standard)
 * - Maximum 128 characters (prevent DoS)
 */
export const passwordSchema = z
    .string()
    .min(8, 'Password must be at least 8 characters')
    .max(128, 'Password is too long');

/**
 * Sign-in form schema
 */
export const signInSchema = z.object({
    email: emailSchema,
    password: passwordSchema,
    rememberMe: z.boolean().optional().default(false),
});

/**
 * Forgot password schema
 */
export const forgotPasswordSchema = z.object({
    email: emailSchema,
});

/**
 * Username validation schema
 * - Minimum 3 characters
 * - Maximum 50 characters
 * - Only alphanumeric, underscore, and hyphen allowed
 * - No spaces
 */
export const usernameSchema = z
    .string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username is too long')
    .regex(/^[a-zA-Z0-9_-]+$/, 'Username can only contain letters, numbers, underscores, and hyphens');

/**
 * Full name validation schema
 * - Optional field
 * - Maximum 255 characters
 * - Basic name characters only
 */
export const fullNameSchema = z
    .string()
    .max(255, 'Name is too long')
    .regex(/^[a-zA-Z\s'-]+$/, 'Name contains invalid characters')
    .optional()
    .or(z.literal(''));

/**
 * Registration form schema
 * Includes password confirmation validation
 */
export const registerSchema = z
    .object({
        email: emailSchema,
        username: usernameSchema,
        password: passwordSchema,
        confirmPassword: z.string().min(1, 'Please confirm your password'),
        fullName: fullNameSchema,
        agreeTerms: z.boolean().refine((val) => val === true, {
            message: 'You must agree to the terms and conditions',
        }),
    })
    .refine((data) => data.password === data.confirmPassword, {
        message: 'Passwords do not match',
        path: ['confirmPassword'],
    });

/**
 * Type exports for form data
 */
export type SignInFormData = z.infer<typeof signInSchema>;
export type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>;
export type RegisterFormData = z.infer<typeof registerSchema>;

/**
 * Password strength calculator
 * Returns: 0 (very weak) to 4 (very strong)
 */
export function calculatePasswordStrength(password: string): number {
    let strength = 0;

    if (password.length >= 8) strength++;
    if (password.length >= 12) strength++;
    if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength++;
    if (/\d/.test(password)) strength++;
    if (/[^a-zA-Z0-9]/.test(password)) strength++;

    return Math.min(strength, 4);
}

/**
 * Get password strength label
 */
export function getPasswordStrengthLabel(strength: number): string {
    const labels = ['Very Weak', 'Weak', 'Fair', 'Good', 'Strong'];
    return labels[strength] || 'Very Weak';
}

/**
 * Get password strength color
 */
export function getPasswordStrengthColor(strength: number): string {
    const colors = ['#ef4444', '#f97316', '#eab308', '#22c55e', '#10b981'];
    return colors[strength] || '#ef4444';
}
