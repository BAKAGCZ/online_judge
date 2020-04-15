from django.contrib.auth.models import AbstractUser
from django.db import models


class User(AbstractUser):
    point = models.IntegerField(default=0, blank=True)
    attempt_number = models.IntegerField(default=0, blank=True)
    solved_number = models.IntegerField(default=0, blank=True)

    class Meta:
        db_table = 'auth_user'
