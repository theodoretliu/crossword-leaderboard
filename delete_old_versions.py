import boto3
import datetime

BUCKET = "crossword-sqlite-backup"
BACKUP_WINDOW = datetime.timedelta(days=30)
BACKUP_PREFIX = "production.sqlite3"
BACKUP_FILES = [
    "production.sqlite3",
    "production.sqlite3-wal",
    "production.sqlite3-shm",
]
FILES_TO_KEEP = ["final_dump.sql"] + BACKUP_FILES

now = datetime.datetime.now(datetime.timezone.utc)

s3 = boto3.resource("s3")
bucket = s3.Bucket(BUCKET)

num_deleted = 0

total_objects = 0
total_size = 0

for obj in bucket.object_versions.all():
    total_objects += 1

    total_size += obj.size

    if obj.object_key not in FILES_TO_KEEP or (
        now - obj.last_modified > BACKUP_WINDOW and obj.object_key in BACKUP_FILES
    ):
        print("deleting", obj)
        obj.delete()
        num_deleted += 1

print('total objects', total_objects)
print('total size (bytes, gigabytes)', total_size, total_size / (1024 ** 3))

print("deleted", num_deleted, "files")
