#!/bin/bash
# 换Anthropic API key后，把main的key同步到所有其他agent
# Usage: bash scripts/sync-api-key.sh

python3 -c "
import json, os

OPENCLAW_DIR = os.path.expanduser('~/.openclaw/agents')
main_path = f'{OPENCLAW_DIR}/main/agent/auth-profiles.json'

with open(main_path) as f:
    main_profiles = json.load(f)

ant_profile = main_profiles.get('profiles', {}).get('anthropic:default', {})
new_key_suffix = ant_profile.get('token', '')[-20:]
print(f'Syncing key ...{new_key_suffix} to all agents')

for ag in os.listdir(OPENCLAW_DIR):
    if ag == 'main': continue
    path = f'{OPENCLAW_DIR}/{ag}/agent/auth-profiles.json'
    try:
        with open(path) as f:
            d = json.load(f)
        old_key = d.get('profiles', {}).get('anthropic:default', {}).get('token', '')
        if old_key == ant_profile.get('token', ''):
            print(f'  {ag}: already up-to-date')
            continue
        d.setdefault('profiles', {})['anthropic:default'] = ant_profile
        with open(path, 'w') as f:
            json.dump(d, f, indent=2)
        print(f'  {ag}: updated')
    except FileNotFoundError:
        pass
    except Exception as e:
        print(f'  {ag}: ERROR - {e}')

print('Done.')
"
