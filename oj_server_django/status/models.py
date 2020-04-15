from django.db import models
import django.utils.timezone as timezone


class Solution(models.Model):
    username = models.CharField(max_length=50)
    problem_id = models.IntegerField()
    result = models.CharField(max_length=36, blank=True)
    memory = models.CharField(max_length=36, blank=True)
    time = models.CharField(max_length=36, blank=True)
    lang = models.CharField(max_length=36)
    length = models.CharField(max_length=36, blank=True)
    submitted = models.DateTimeField(default=timezone.now, blank=True)
    code = models.TextField()
