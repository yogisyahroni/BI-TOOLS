import re
import sys

# Read the file
with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_organization_handler.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Define replacements for each specific error
replacements = [
    (r'"error": err\.Error\(\),\s*\/\/ Line 244 context: AddOrganizationMember', '"error": "Failed to add member",'),
    (r'"error": err\.Error\(\),\s*\/\/ Line 266', '"error": "Failed to retrieve members",'),
    (r'"error": err\.Error\(\),\s*\/\/ Line 296', '"error": "Failed to remove member",'),
    (r'"error": err\.Error\(\),\s*\/\/ Line 317', '"error": "Failed to update member role",'),
]

# Count occurrences
occurrences = re.findall(r'"error": err\.Error\(\)', content)
print(f"Found {len(occurrences)} occurrences of err.Error()")

# Simple approach: Find each occurrence in context and replace
lines = content.split('\n')
result_lines = []
for i, line in enumerate(lines, 1):
    if '"error": err.Error(),' in line:
        # Determine context based on nearby lines
        context_start = max(0, i-10)
        context = '\n'.join(lines[context_start:i+5])
        
        if 'GetOrganizationMembers' in context or i == 266:
            line = line.replace('err.Error()', '"Failed to retrieve members"')
        elif 'AddOrganizationMember' in context or i == 244:
            line = line.replace('err.Error()', '"Failed to add member"')
        elif 'RemoveOrganizationMember' in context or i == 296:
            line = line.replace('err.Error()', '"Failed to remove member"')
        elif 'UpdateMemberRole' in context or i == 317:
            line = line.replace('err.Error()', '"Failed to update member role"')
    
    result_lines.append(line)

# Write back
with open(r'e:\antigraviti google\inside engine\insight-engine-ai-ui\backend\handlers\admin_organization_handler.go', 'w', encoding='utf-8') as f:
    f.write('\n'.join(result_lines))

print("Fixed all error exposures!")
