# Authentication Wireframes

## Overview

Authentication flow screens including Login, Register, Forgot Password, and 2FA.

---

## Login Page (Split Layout)

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                     │
│  ┌─────────────────────────────────┐  ┌─────────────────────────────────────────┐  │
│  │                                 │  │                                         │  │
│  │         [NexusOS Logo]          │  │                                         │  │
│  │                                 │  │                                         │  │
│  │                                 │  │         ┌───────────────────────┐       │  │
│  │       Welcome back              │  │         │                       │       │  │
│  │       Sign in to continue       │  │         │    3D Abstract Art    │       │  │
│  │                                 │  │         │    Gradient BG        │       │  │
│  │                                 │  │         │    + Noise Texture    │       │  │
│  │  Email                          │  │         │                       │       │  │
│  │  ┌─────────────────────────┐   │  │         │    (Indigo → Violet)  │       │  │
│  │  │ john@example.com        │   │  │         │                       │       │  │
│  │  └─────────────────────────┘   │  │         └───────────────────────┘       │  │
│  │                                 │  │                                         │  │
│  │  Password                       │  │                                         │  │
│  │  ┌─────────────────────────┐   │  │    "Trusted by 500+ companies           │  │
│  │  │ ●●●●●●●●           [👁] │   │  │     worldwide"                          │  │
│  │  └─────────────────────────┘   │  │                                         │  │
│  │                                 │  │    ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐  │  │
│  │  ┌────────────────────────────┐│  │    │Logo│ │Logo│ │Logo│ │Logo│ │Logo│  │  │
│  │  │ ☐ Remember me              ││  │    └────┘ └────┘ └────┘ └────┘ └────┘  │  │
│  │  │              Forgot password││  │    ↑ Partner logos (grayscale)         │  │
│  │  └────────────────────────────┘│  │                                         │  │
│  │                                 │  │                                         │  │
│  │  ┌─────────────────────────┐   │  │                                         │  │
│  │  │       Sign In            │   │  │                                         │  │
│  │  └─────────────────────────┘   │  │                                         │  │
│  │          Primary Button         │  │                                         │  │
│  │                                 │  │                                         │  │
│  │  ─────────── or ───────────    │  │                                         │  │
│  │                                 │  │                                         │  │
│  │  ┌─────────────────────────┐   │  │                                         │  │
│  │  │ [G] Sign in with Google  │   │  │                                         │  │
│  │  └─────────────────────────┘   │  │                                         │  │
│  │                                 │  │                                         │  │
│  │  ┌─────────────────────────┐   │  │                                         │  │
│  │  │ [MS] Sign in with Azure  │   │  │                                         │  │
│  │  └─────────────────────────┘   │  │                                         │  │
│  │                                 │  │                                         │  │
│  │                                 │  │                                         │  │
│  │  Don't have an account?         │  │                                         │  │
│  │  Create account                 │  │                                         │  │
│  │                                 │  │                                         │  │
│  └─────────────────────────────────┘  └─────────────────────────────────────────┘  │
│                                                                                     │
│            LEFT PANEL (50%)                      RIGHT PANEL (50%)                 │
│            Functional Zone                       Visual/Branding Zone              │
│                                                  (Hidden on mobile)                │
│                                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Register Page

```
┌───────────────────────────────────────┬─────────────────────────────────────────┐
│                                       │                                         │
│         [NexusOS Logo]                │                                         │
│                                       │                                         │
│                                       │         ┌───────────────────────┐       │
│       Create your account             │         │                       │       │
│       Get started in minutes          │         │    Dashboard Preview  │       │
│                                       │         │    Screenshot (Skewed)│       │
│                                       │         │                       │       │
│  Full Name                            │         │    "See what you can  │       │
│  ┌───────────────────────────────┐   │         │     build with NexusOS"│       │
│  │ John Doe                      │   │         │                       │       │
│  └───────────────────────────────┘   │         └───────────────────────┘       │
│                                       │                                         │
│  Work Email                           │                                         │
│  ┌───────────────────────────────┐   │         Features:                        │
│  │ john@company.com              │   │         ✓ Role-based access control     │
│  └───────────────────────────────┘   │         ✓ Real-time audit logging       │
│                                       │         ✓ AI-powered assistance         │
│  Password                             │         ✓ Multi-tenant ready            │
│  ┌───────────────────────────────┐   │                                         │
│  │ ●●●●●●●●                 [👁] │   │                                         │
│  └───────────────────────────────┘   │                                         │
│  ■■■■░░░░░░░░░░░░  Weak               │                                         │
│  ⚠️ Add numbers and symbols           │                                         │
│                                       │                                         │
│  Confirm Password                     │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │ ●●●●●●●●                 [👁] │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  ☑ I agree to the Terms of Service   │                                         │
│    and Privacy Policy                 │                                         │
│                                       │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │       Create Account           │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  Already have an account? Sign in     │                                         │
│                                       │                                         │
└───────────────────────────────────────┴─────────────────────────────────────────┘
```

---

## Forgot Password

```
┌───────────────────────────────────────┬─────────────────────────────────────────┐
│                                       │                                         │
│         [NexusOS Logo]                │                                         │
│                                       │                                         │
│                                       │                                         │
│       Forgot your password?           │         ┌───────────────────────┐       │
│       No worries, we'll send you      │         │                       │       │
│       reset instructions.             │         │    🔐                  │       │
│                                       │         │    Lock illustration  │       │
│                                       │         │                       │       │
│  Email                                │         └───────────────────────┘       │
│  ┌───────────────────────────────┐   │                                         │
│  │ john@example.com              │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │       Send Reset Link          │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  [← Back to login]                    │                                         │
│                                       │                                         │
└───────────────────────────────────────┴─────────────────────────────────────────┘
```

---

## Email Sent Confirmation

```
┌───────────────────────────────────────┬─────────────────────────────────────────┐
│                                       │                                         │
│         [NexusOS Logo]                │                                         │
│                                       │                                         │
│                                       │                                         │
│           ┌───────────┐               │         ┌───────────────────────┐       │
│           │    ✉️     │               │         │                       │       │
│           └───────────┘               │         │    ✈️                  │       │
│                                       │         │    Email flying       │       │
│       Check your email                │         │    illustration       │       │
│                                       │         │                       │       │
│       We sent a password reset link   │         └───────────────────────┘       │
│       to john@example.com             │                                         │
│                                       │                                         │
│       Didn't receive the email?       │                                         │
│       Check spam folder or            │                                         │
│       [Resend email]                  │                                         │
│                                       │                                         │
│       [← Back to login]               │                                         │
│                                       │                                         │
└───────────────────────────────────────┴─────────────────────────────────────────┘
```

---

## Reset Password

```
┌───────────────────────────────────────┬─────────────────────────────────────────┐
│                                       │                                         │
│         [NexusOS Logo]                │                                         │
│                                       │                                         │
│                                       │                                         │
│       Set new password                │         ┌───────────────────────┐       │
│       Create a strong password        │         │                       │       │
│       for your account                │         │    🔑                  │       │
│                                       │         │    Key illustration   │       │
│                                       │         │                       │       │
│  New Password                         │         └───────────────────────┘       │
│  ┌───────────────────────────────┐   │                                         │
│  │ ●●●●●●●●●●●●            [👁] │   │                                         │
│  └───────────────────────────────┘   │                                         │
│  ■■■■■■■■■■■■░░░░  Strong ✓          │                                         │
│                                       │                                         │
│  Confirm Password                     │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │ ●●●●●●●●●●●●            [👁] │   │                                         │
│  └───────────────────────────────┘   │                                         │
│  ✓ Passwords match                   │                                         │
│                                       │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │       Reset Password           │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  [← Back to login]                    │                                         │
│                                       │                                         │
└───────────────────────────────────────┴─────────────────────────────────────────┘
```

---

## Two-Factor Authentication

```
┌───────────────────────────────────────┬─────────────────────────────────────────┐
│                                       │                                         │
│         [NexusOS Logo]                │                                         │
│                                       │                                         │
│                                       │                                         │
│       Two-factor authentication       │         ┌───────────────────────┐       │
│       Enter the code from your        │         │                       │       │
│       authenticator app               │         │    📱                  │       │
│                                       │         │    Phone with code    │       │
│                                       │         │    illustration       │       │
│  ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐ │         │                       │       │
│  │ 1 │ │ 2 │ │ 3 │ │ 4 │ │ _ │ │ _ │ │         └───────────────────────┘       │
│  └───┘ └───┘ └───┘ └───┘ └───┘ └───┘ │                                         │
│           ↑ 6-digit code entry        │                                         │
│                                       │                                         │
│  ┌───────────────────────────────┐   │                                         │
│  │       Verify Code              │   │                                         │
│  └───────────────────────────────┘   │                                         │
│                                       │                                         │
│  Can't access authenticator?          │                                         │
│  [Use backup code]                    │                                         │
│                                       │                                         │
│  [← Back to login]                    │                                         │
│                                       │                                         │
└───────────────────────────────────────┴─────────────────────────────────────────┘
```

---

## Mobile Layout (< 768px)

```
┌──────────────────────────┐
│                          │
│      [NexusOS Logo]      │
│                          │
│    Welcome back          │
│    Sign in to continue   │
│                          │
│  Email                   │
│  ┌────────────────────┐  │
│  │ john@example.com   │  │
│  └────────────────────┘  │
│                          │
│  Password                │
│  ┌────────────────────┐  │
│  │ ●●●●●●●●      [👁] │  │
│  └────────────────────┘  │
│                          │
│  ☐ Remember me           │
│  Forgot password?        │
│                          │
│  ┌────────────────────┐  │
│  │     Sign In        │  │
│  └────────────────────┘  │
│                          │
│  ──────── or ────────   │
│                          │
│  ┌────────────────────┐  │
│  │ [G] Google         │  │
│  └────────────────────┘  │
│                          │
│  ┌────────────────────┐  │
│  │ [MS] Microsoft     │  │
│  └────────────────────┘  │
│                          │
│  Don't have an account?  │
│  Create account          │
│                          │
└──────────────────────────┘

Note: Right panel (branding) is hidden
on mobile to maximize form space
```

---

## Error States

### Invalid Credentials

```
  Email
  ┌───────────────────────────────┐
  │ john@example.com              │
  └───────────────────────────────┘

  Password
  ┌───────────────────────────────┐
  │ ●●●●●●●●                 [👁] │
  └───────────────────────────────┘
  ⚠️ Invalid email or password. Please try again.
     [Forgot your password?]
```

### Account Locked

```
  ┌─────────────────────────────────────────────────┐
  │ ⚠️ Account Temporarily Locked                   │
  │                                                 │
  │ Too many failed login attempts.                 │
  │ Please try again in 15 minutes or              │
  │ reset your password.                           │
  │                                                 │
  │ [Reset Password]   [Contact Support]           │
  └─────────────────────────────────────────────────┘
```

### Session Expired

```
  ┌─────────────────────────────────────────────────┐
  │ ℹ️ Session Expired                              │
  │                                                 │
  │ Your session has expired for security reasons.  │
  │ Please sign in again to continue.              │
  │                                                 │
  │                              [Sign In Again]   │
  └─────────────────────────────────────────────────┘
```
