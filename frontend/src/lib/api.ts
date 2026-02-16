import type { ErrorResponse } from './types';

const BASE_URL = 'http://localhost:8080/api/v1';

export async function apiFetch<T>(
    path: string,
    method: 'GET' | 'POST' = 'GET',
    body: unknown = null
): Promise<T> {
    const options: RequestInit = {
        method,
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include'
    };

    if (body) {
        options.body = JSON.stringify(body);
    }

    const res = await fetch(`${BASE_URL}${path}`, options);

    // Check if response is empty (like your 201/200 might be before parsing)
    const isJson = res.headers.get('content-type')?.includes('application/json');
    const data = isJson ? await res.json() : null;

    if (!res.ok) {
        throw data as ErrorResponse;
    }

    return data as T;
}