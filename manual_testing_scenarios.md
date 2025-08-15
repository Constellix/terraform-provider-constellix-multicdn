# Manual Test Scenarios & Test Cases – Terraform + Backend CRUD APIs

---

## Scenario 1: Create Configuration Resource

### Test Case 1.1: Successful creation

**Steps:** (Perform these steps for both cdn config and preference config)
1. Write a valid `.tf` file with resource block for configuration.
2. Run `terraform apply`.
3. Verify:
    - Terraform shows successful creation.
    - Resource exists in backend (via API or UI).
    - Terraform state file reflects the new resource.

### Test Case 1.2: Backend API fails during creation

**Steps:**
1. Simulate backend `POST` failure (e.g., return 500 or 400).
2. Run `terraform apply`.
3. Verify:
    - Terraform fails gracefully with an appropriate error.
    - No resource is created on the backend.
    - Terraform state is unchanged.

---

## Scenario 2: Read Configuration Resource

### Test Case 2.1: Resource exists in backend and state is correct

**Steps:**
1. Apply the resource.
2. Run `terraform refresh`.
3. Verify: No drift is detected.

### Test Case 2.2: Resource deleted manually in backend

**Steps:**
1. Create resource using Terraform.
2. Delete the resource directly from the backend.
3. Run `terraform plan`.
4. Verify: Terraform marks the resource as “to be created” (shows drift).

---

## Scenario 3: Update Configuration Resource

### Test Case 3.1: Update attribute value

**Steps:**
1. Create an initial resource.
2. Modify the `.tf` file (e.g., change a property value).
3. Run `terraform apply`.
4. Verify:
    - Backend reflects the updated value.
    - State file is updated.
    - No additional resources are created or deleted.

### Test Case 3.2: Update fails in backend

**Steps:**
1. Simulate backend `PUT` failure.
2. Apply the change using Terraform.
3. Verify:
    - Terraform reports the failure.
    - Retries or fallback behavior is handled appropriately.

---

## Scenario 4: Delete Configuration Resource

### Test Case 4.1: Successful deletion

**Steps:**
1. Create and verify the resource.
2. Remove the resource block from the `.tf` file.
3. Run `terraform apply`.
4. Verify:
    - Resource is deleted in the backend.
    - Resource is removed from the Terraform state.

### Test Case 4.2: Backend fails during delete

**Steps:**
1. Simulate backend `DELETE` failure (e.g., 500 or timeout).
2. Run `terraform apply`.
3. Verify:
    - Terraform reports the error.
    - Resource remains in state and is not falsely removed.

---

## Scenario 5: Idempotency and Drift Detection

### Test Case 5.1: Re-apply without changes (idempotency check)

**Steps:**
1. Create and apply the resource.
2. Run `terraform apply` again without modifying `.tf` file.
3. Verify: No changes are applied.

### Test Case 5.2: Manual drift in backend

**Steps:**
1. Apply config via Terraform.
2. Modify backend config manually.
3. Run `terraform plan`.
4. Verify: Drift is detected and displayed in the plan.

---

## Scenario 6: Terraform State File Validation

### Test Case 6.1: State file reflects backend correctly

**Steps:**
1. Apply the resource using Terraform.
2. Inspect the `.tfstate` file.
3. Verify that values match the current state in the backend.

---

## Scenario 7: Error and Edge Case Handling

### Test Case 7.1: Missing required fields in `.tf`

**Steps:**
1. Create a `.tf` file with missing required attributes.
2. Run `terraform plan` or `apply`.
3. Verify:
    - Terraform shows a validation error.
    - No resource is created or changed.

### Test Case 7.2: Invalid field values

**Steps:**
1. Use invalid data types or enums in the `.tf` file.
2. Run `terraform apply`.
3. Verify:
    - Terraform or backend rejects the request with a clear error.

---

## Optional Integration & Functional Test Ideas

- Use `terraform plan` diff to inspect changes before applying.
- Validate resource existence and correctness via API clients (e.g., Postman or curl).
- Test with multiple resources in one `.tf` file.
- Use `terraform workspace` to test isolated environments.
- Run `terraform destroy` to test cleanup behavior.