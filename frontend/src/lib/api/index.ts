import { PUBLIC_BASE_URL } from '$env/static/public';
import { dev } from '$app/environment';

interface FetchOptions extends Omit<RequestInit, 'body' | 'method'> {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
    body?: any;
    fetcher?: typeof fetch;
}

/**
 * @throws {ErrorResponse & { status: number }} 
 */
export async function apiFetch<T>(path: string, opts?: FetchOptions, fetcher: typeof fetch = fetch): Promise<T> {
    opts = {
        ...opts,
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        method: opts?.method ?? 'GET',
        body: opts?.body && JSON.stringify(opts.body),
    };

    const response = await fetcher(`${PUBLIC_BASE_URL}${path}`, opts);

    const contentType = response.headers.get('content-type');
    let responseBody;

    if (response.status === 204) { }
    else if (contentType?.includes('application/json')) {
        responseBody = await response.json();
    } else {
        responseBody = await response.text();
    }

    if (dev) {
        const color = response.ok ? 'color: #10b981' : 'color: #ef4444';

        console.groupCollapsed(`✅ API Res: [${response.status}] ${path}`);
        console.log('%cContent-Type:', color, contentType);
        console.log('%cStatus:', color, response.status, response.statusText);
        console.log('Request Body:', opts.body);
        console.log('Response Body:', responseBody);
        console.groupEnd();
    }

    if (response.ok) {
        return responseBody;
    }

    if (contentType?.includes('application/json')) {
        throw { status: response.status, ...responseBody };
    }

    throw { status: response.status, details: responseBody };
};