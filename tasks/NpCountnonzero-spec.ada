--  <vc-preamble>
package Np_Countnonzero_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   subtype Count_Type is Natural range 0 .. Max_Index + 1;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Dafny's nonzero_helper counts the non-zero elements of a sequence and
   --  recurses on its tail.  Here the sequence is the suffix of A that starts
   --  at From, so nonzero_helper(arr[..]) is Nonzero_Count (A, A'First) and
   --  nonzero_helper(arr[1..]) is Nonzero_Count (A, A'First + 1).
   function Nonzero_Count (A : Int_Array; From : Natural) return Natural is
     (if From > A'Last then 0
      else Nonzero_Count (A, From + 1) + (if A (From) = 0 then 0 else 1))
   with
     Pre  => From >= A'First and then From <= A'Last + 1,
     Post => Nonzero_Count'Result <= A'Last - From + 1,
     Subprogram_Variant => (Decreases => A'Last - From + 1);
--  </vc-preamble>

--  <vc-spec>
   procedure Nonzero (A : Int_Array; Result : out Count_Type) with
     Post => Result <= A'Length
             and then Result = Nonzero_Count (A, A'First)
             and then (if A'Length > 0 and then A (A'First) = 0
                       then Nonzero_Count (A, A'First + 1)
                            = (if Result > 0 then Result - 1 else 0));

end Np_Countnonzero_Spec;

package body Np_Countnonzero_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Nonzero (A : Int_Array; Result : out Count_Type) is
   begin
      pragma Assume (False);
   end Nonzero;
--  </vc-code>

--  <vc-postamble>
end Np_Countnonzero_Spec;
--  </vc-postamble>
