from django.db import models


class Problem(models.Model):
    title = models.CharField(max_length=255)
    time_limit = models.CharField(max_length=36)
    memory_limit = models.CharField(max_length=36)
    point = models.IntegerField()
    description = models.TextField()
    input_format = models.TextField(blank=True)
    output_format = models.TextField(blank=True)
    sample_input = models.TextField(blank=True)
    sample_output = models.TextField(blank=True)
    hint = models.TextField(blank=True)
    source = models.CharField(max_length=100, blank=True)

    def __str__(self):
        return self.title
