from locust import HttpUser, task, between
import json
import random

class FlashSaleUserAsync(HttpUser):
    def on_start(self):
        self.customer_id = random.randint(1000, 9999)
        
    @task
    def place_order_async(self):
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "item_id": f"item-{random.randint(1, 100)}",
                    "name": f"Product {random.randint(1, 100)}",
                    "price": round(random.uniform(10, 200), 2),
                    "quantity": random.randint(1, 5)
                }
            ]
        }
        
        with self.client.post("/orders/async", 
                              json=order_data,
                              catch_response=True) as response:
            if response.status_code in [200, 202]:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")