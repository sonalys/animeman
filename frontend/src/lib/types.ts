export interface UserRegistration {
    username: string;
    password: string;
}

export interface AuthResponse {
    userID: string;
}

export interface FieldError {
    field: string;
    code: 'alreadyExists' | 'minLength' | 'maxLength' | 'required' | 'invalidFormat' | 'unknown';
    message: string;
}

export interface ErrorResponse {
    traceID?: string;
    details?: string;
    fieldErrors?: FieldError[];
}

export interface UserRegistration {
    username: string; // pattern: ^[a-zA-Z0-9_]+$
    password: string; // min: 8, max: 72
}