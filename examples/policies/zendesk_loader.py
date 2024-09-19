import asyncio
import json
import os

import httpx

ZENDESK_SUBDOMAIN = os.environ["ZENDESK_SUBDOMAIN"]
ZENDESK_EMAIL = os.environ["ZENDESK_EMAIL"]
ZENDESK_API_TOKEN = os.environ["ZENDESK_API_TOKEN"]

auth = httpx.BasicAuth(ZENDESK_EMAIL, ZENDESK_API_TOKEN)
base_url = f"https://{ZENDESK_SUBDOMAIN}.zendesk.com"

users_cache: dict[str, dict | None] = {}


async def get_user(user_id: str) -> dict | None:
    if user := users_cache.get(user_id):
        return user

    async with httpx.AsyncClient(base_url=base_url, auth=auth) as client:
        response = await client.get(f"/api/v2/users/{user_id}.json")

    response.raise_for_status()
    user = response.json()["user"]
    users_cache[user_id] = user
    return user


async def get_tickets() -> list[dict]:
    async with httpx.AsyncClient(base_url=base_url, auth=auth) as client:
        params = {"query": "status:new status:open status:pending"}
        response = await client.get("/api/v2/search.json", params=params)

    response.raise_for_status()
    return response.json()["results"]


async def main():
    tickets = await get_tickets()

    for ticket in tickets:
        if requester_id := ticket.get("requester_id"):
            ticket["requester"] = await get_user(requester_id)
        if submitter_id := ticket.get("submitter_id"):
            ticket["submitter"] = await get_user(submitter_id)
        if assignee_id := ticket.get("assignee_id"):
            ticket["assignee"] = await get_user(assignee_id)

        ticket["url"] = f"{base_url}/agent/tickets/{ticket['id']}"

    print(json.dumps(tickets))


asyncio.run(main())
