import boto3
from tqdm import tqdm

BUCKETS = ['crossword-sqlite-litestream', 'crossword-sqlite-backup']

s3 = boto3.resource('s3')

total_cost = 0

for bucket in BUCKETS:
    b = s3.Bucket(bucket)

    count = 0
    total_size = 0

    for obj in tqdm(b.object_versions.all()):
        count += 1
        if obj.size is not None:
            total_size += obj.size


    size_gb = total_size / (1024 ** 3)

    total_cost += size_gb * 0.023

    print('bucket:', bucket)
    print('num objects:', count)
    print('size (GB)', size_gb)
    print()

print('estimated monthly cost:', total_cost)
