import { randomBytes, createHash } from 'crypto';
import { supabase, API_KEY_TABLE } from '../config/config.js';


const checkKeyExist = async () => {
    const { data, error } = await supabase
        .from(API_KEY_TABLE)
        .select('id')
        .limit(1);

    if (error) {
        return false;
    }
    return data && data.length > 0;
}

export async function generateApiKey(name: string, description?: string) {
    if (await checkKeyExist()) {
        throw new Error('An API key already exists. Multiple keys are not supported in this version. Delete the existing key through the supabase dashboard to create a new one.');
    }

    const apiKey = `axh_${randomBytes(100).toString('hex')}`;

    const hashedKey = createHash('sha256').update(apiKey).digest('hex');

    // Store in database
    const { data, error } = await supabase
        .from(API_KEY_TABLE)
        .insert({
            name,
            description,
            key_hash: hashedKey,
            key_prefix: apiKey.substring(0, 12),
            is_active: true,
            created_at: new Date().toISOString()
        })
        .select()
        .single();

    if (error) {
        console.error('Database error:', error);
        throw new Error('Failed to store API key');
    }

    return {
        id: data.id,
        api_key: apiKey,
        name: data.name,
        key_prefix: data.key_prefix,
        created_at: data.created_at,
        warning: 'Store this API key securely. It will not be shown again.'
    };
}

export async function getApiKeyInfo(keyId: string) {
    const { data, error } = await supabase
        .from(API_KEY_TABLE)
        .select('id, name, description, key_prefix, is_active, created_at, last_used_at, usage_count')
        .eq('id', keyId)
        .single();

    if (error) return null;
    return data;
}

export async function updateKeyUsage(keyHash: string) {
    const { error } = await supabase.rpc("increment_key_usage", {
        key_hash: keyHash
    });

    if (error) {
        console.error("Failed to update key usage:", error);
    }
}
