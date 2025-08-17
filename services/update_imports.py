#!/usr/bin/env python3
import os
import re
import glob

def update_imports_in_file(filepath):
    """Update import paths in a single Go file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        # Replace the import paths
        updated_content = content.replace(
            'github.com/lendingplatform/los/services',
            'github.com/huuhoait/los-demo/services'
        )
        
        # Only write if content changed
        if updated_content != content:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(updated_content)
            print(f"Updated: {filepath}")
            return True
        return False
    except Exception as e:
        print(f"Error updating {filepath}: {e}")
        return False

def main():
    # Find all Go files
    go_files = []
    for root, dirs, files in os.walk('.'):
        for file in files:
            if file.endswith('.go'):
                go_files.append(os.path.join(root, file))
    
    updated_count = 0
    for go_file in go_files:
        if update_imports_in_file(go_file):
            updated_count += 1
    
    print(f"Updated {updated_count} files")

if __name__ == "__main__":
    main()
